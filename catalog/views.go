package catalog

import (
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

const (
	DirectoryTerminationValue = 1_000_000
	DirectoryTermination      = string(rune(DirectoryTerminationValue))
)

func sqEntriesV(requestedCommit CommitID) sq.SelectBuilder {
	entriesQ := sq.Select("*",
		"min_commit > 0 AS is_committed",
		"max_commit = 0 AS is_tombstone",
		"ctid AS entry_ctid\n",
		"max_commit < max_commit_id() AS is_deleted",
		"CASE WHEN min_commit = 0 THEN max_commit_id() ELSE min_commit END AS commit_weight").
		From("entries")
	switch requestedCommit {
	case UncommittedID: // no further filtering is required
	case CommittedID:
		entriesQ = sq.Select("*").FromSelect(entriesQ, "t2").Where("is_committed")
	default:
		entriesQ = sq.Select("*").FromSelect(entriesQ, "t2").Where("? >=  min_commit and is_committed", requestedCommit)
	}
	return entriesQ
}

func sqEntriesLineage(branchID int64, requestedCommit CommitID, lineage []lineageCommit) sq.SelectBuilder {
	isDisplayedBranch := "e.branch_id = " + strconv.FormatInt(branchID, 10)
	maxCommitExpr := sq.Case().When(isDisplayedBranch, "e.max_commit\n")
	isDeletedExpr := sq.Case().When(isDisplayedBranch, "e.is_deleted\n")
	lineageFilter := "(" + isDisplayedBranch + ")\n"
	for _, lc := range lineage {
		branchCond := "e.branch_id = " + strconv.FormatInt(lc.BranchID, 10)
		commitStr := strconv.FormatInt(int64(lc.CommitID), 10)
		ancestorCond := branchCond + " and e.max_commit < " + commitStr
		maxCommitExpr = maxCommitExpr.When(ancestorCond, "e.max_commit\n")
		isDeletedExpr = isDeletedExpr.When(ancestorCond, "e.is_deleted\n")
		lineageFilter += " OR (" + branchCond + " AND e.min_commit <= " + commitStr + " AND e.is_committed) \n"
	}
	maxCommitExpr = maxCommitExpr.Else("max_commit_id()")
	maxCommitAlias := sq.Alias(maxCommitExpr, "max_commit")
	isDeletedExpr = isDeletedExpr.Else("false")
	isDeletedAlias := sq.Alias(isDeletedExpr, "is_deleted")
	baseSelect := sq.Select().Distinct().Options(" ON (e.path) ").
		FromSelect(sqEntriesV(requestedCommit), "e\n").
		Where(lineageFilter).
		OrderBy("e.path", "source_branch desc", "e.commit_weight desc").
		Columns(strconv.FormatInt(branchID, 10)+" AS displayed_branch",
			"e.path", "e.branch_id AS source_branch",
			"e.min_commit", "e.physical_address",
			"e.creation_date", "e.size", "e.checksum", "e.metadata",
			"e.is_committed", "e.is_tombstone", "e.entry_ctid", "e.is_expired").
		Column(maxCommitAlias).Column(isDeletedAlias)
	return baseSelect
}

func sqLineageConditions(branchID int64, lineage []lineageCommit) (string, sq.Sqlizer, sq.Sqlizer) {
	isDisplayedBranch := "e.branch_id = " + strconv.FormatInt(branchID, 10)
	maxCommitExpr := sq.Case().When(isDisplayedBranch, "e.max_commit\n")
	isDeletedExpr := sq.Case().When(isDisplayedBranch, "e.is_deleted\n")
	lineageFilter := "(" + isDisplayedBranch + ")\n"
	for _, lc := range lineage {
		branchCond := "e.branch_id = " + strconv.FormatInt(lc.BranchID, 10)
		commitStr := strconv.FormatInt(int64(lc.CommitID), 10)
		ancestorCond := branchCond + " and e.max_commit < " + commitStr
		maxCommitExpr = maxCommitExpr.When(ancestorCond, "e.max_commit\n")
		isDeletedExpr = isDeletedExpr.When(ancestorCond, "e.is_deleted\n")
		lineageFilter += " OR (" + branchCond + " AND e.min_commit <= " + commitStr + " AND e.is_committed) \n"
	}
	maxCommitExpr = maxCommitExpr.Else("max_commit_id()")
	maxCommitAlias := sq.Alias(maxCommitExpr, "max_commit")
	isDeletedExpr = isDeletedExpr.Else("false")
	isDeletedAlias := sq.Alias(isDeletedExpr, "is_deleted")
	return lineageFilter, maxCommitAlias, isDeletedAlias
}

func sqEntriesLineageV(branchID int64, requestedCommit CommitID, lineage []lineageCommit) sq.SelectBuilder {
	lineageFilter,
		maxCommitAlias,
		isDeletedAlias := sqLineageConditions(branchID, lineage)
	baseSelect := sq.Select().Distinct().Options(" ON (e.path) ").
		FromSelect(sqEntriesV(requestedCommit), "e\n").
		Where(lineageFilter).
		OrderBy("e.path", "source_branch desc", "e.commit_weight desc").
		Column("? AS displayed_branch", strconv.FormatInt(branchID, 10)).
		Columns("e.path", "e.branch_id AS source_branch",
			"e.min_commit", "e.physical_address",
			"e.creation_date", "e.size", "e.checksum", "e.metadata",
			"e.is_committed", "e.is_tombstone", "e.entry_ctid", "e.is_expired").
		Column(maxCommitAlias).Column(isDeletedAlias)
	return baseSelect
}

func sqDiffFromSonV(fatherID, sonID int64, fatherEffectiveCommit, sonEffectiveCommit CommitID, fatherUncommittedLineage []lineageCommit, sonLineageValues string) sq.SelectBuilder {
	lineage := sqEntriesLineage(fatherID, UncommittedID, fatherUncommittedLineage)
	fatherSQL, fatherArgs := sq.Select("*").FromSelect(lineage, "z").
		Where("displayed_branch = ?", fatherID).MustSql()
	// Can diff with expired files, just not usefully!
	fromSonInternalQ := sq.Select("s.path",
		"s.is_deleted AS DifferenceTypeRemoved",
		"f.path IS NOT NULL AND NOT f.is_deleted AS DifferenceTypeChanged",
		"COALESCE(f.is_deleted, true) AND s.is_deleted AS both_deleted",
		"f.path IS NOT NULL AND (f.physical_address = s.physical_address AND f.is_deleted = s.is_deleted) AS same_object",
		"s.entry_ctid",
		"f.source_branch",
	).
		//Conflict detection
		Column(`-- father either created or deleted after last merge  - conflict
			f.path IS NOT NULL AND ( NOT f.is_committed OR -- uncommitted entries always new
									(f.source_branch = ? AND  -- it is the father branch - not from lineage
									( f.min_commit > ? OR -- created after last merge
									 (f.max_commit >= ? AND f.is_deleted))) -- deleted after last merge
									OR (f.source_branch != ? AND  -- an entry from father lineage
				-- negative proof - if the son could see this object - than this is NOT a conflict
				-- done by examining the son lineage against the father object
									 NOT EXISTS ( SELECT * FROM`+sonLineageValues+` WHERE
											l.branch_id = f.source_branch AND
										-- prove that ancestor entry  was observable by the son
											(l.commit_id >= f.min_commit AND
											 (l.commit_id > f.max_commit OR NOT f.is_deleted))
										   ))) 
											AS DifferenceTypeConflict `, fatherID, fatherEffectiveCommit, fatherEffectiveCommit, fatherID).
		FromSelect(sqEntriesV(CommittedID).
			Where("branch_id = ? AND (min_commit >= ? OR max_commit >= ? and is_deleted)", sonID, sonEffectiveCommit, sonEffectiveCommit), "s").
		LeftJoin("("+fatherSQL+") AS f ON f.path = s.path", fatherArgs...)
	RemoveNonRelevantQ := sq.Select("*").FromSelect(fromSonInternalQ, "t").Where("NOT (same_object OR both_deleted)")
	return sq.Select().
		Column(sq.Alias(sq.Case().When("DifferenceTypeConflict", "3").
			When("DifferenceTypeRemoved", "1").
			When("DifferenceTypeChanged", "2").
			Else("0"), "diff_type")).
		Column("path").Column(sq.Alias(sq.Case().
		When("NOT(DifferenceTypeConflict OR DifferenceTypeRemoved)", "entry_ctid").
		Else("NULL"),
		"entry_ctid")).
		Column("source_branch").
		FromSelect(RemoveNonRelevantQ, "t1")
}

func sqDiffFromFatherV(fatherID, sonID int64, lastSonMergeWithFather CommitID, fatherUncommittedLineage, sonUncommittedLineage []lineageCommit) sq.SelectBuilder {
	sonLineageValues := getLineageAsValues(sonUncommittedLineage, sonID)
	sonLineage := sqEntriesLineage(sonID, UncommittedID, sonUncommittedLineage)
	sonSQL, sonArgs := sq.Select("*").FromSelect(sonLineage, "s").
		Where("displayed_branch = ? ", sonID).MustSql()

	fatherLineage := sqEntriesLineage(fatherID, CommittedID, fatherUncommittedLineage)
	// Can diff with expired files, just not usefully!
	internalV := sq.Select("f.path",
		"f.entry_ctid",
		"f.is_deleted AS DifferenceTypeRemoved",
		"s.path IS NOT NULL AND not s.is_deleted AS DifferenceTypeChanged",
		"COALESCE(s.is_deleted, true) AND f.is_deleted AS both_deleted",
		//both point to same object, and have the same deletion status
		"s.path IS NOT NULL AND f.physical_address = s.physical_address AND f.is_deleted = s.is_deleted AS same_object").
		Column(`f.min_commit > l.commit_id  -- father created after commit
			OR f.max_commit >= l.commit_id AND f.is_deleted -- father deleted after commit
									AS father_changed`). // father was changed if son could no "see" it
		// this happens if min_commit is larger than the lineage commit
		// or entry deletion max_commit is larger or equal than lineage commit
		Column("s.path IS NOT NULL AND s.source_branch = ? as entry_in_son", sonID).
		Column(`s.path IS NOT NULL AND s.source_branch = ? AND
							(NOT s.is_committed -- uncommitted is new
							 OR s.min_commit > ? -- created after last merge
                           OR (s.max_commit >= ? AND s.is_deleted)) -- deleted after last merge
						  AS DifferenceTypeConflict`, sonID, lastSonMergeWithFather, lastSonMergeWithFather).
		FromSelect(fatherLineage, "f").
		Where("f.displayed_branch = ?", fatherID).
		LeftJoin("("+sonSQL+") AS s ON f.path = s.path", sonArgs...).
		Join(`(SELECT * FROM ` + sonLineageValues + `) l ON f.source_branch = l.branch_id`)

	RemoveNonRelevantQ := sq.Select("*").
		FromSelect(internalV, "t").
		Where("father_changed AND NOT (same_object OR both_deleted)")

	return sq.Select().
		Column(sq.Alias(sq.Case().When("DifferenceTypeConflict", "3").
			When("DifferenceTypeRemoved", "1").
			When("DifferenceTypeChanged", "2").
			Else("0"), "diff_type")).
		Column("path").
		Column(sq.Alias(sq.Case().
			When("DifferenceTypeChanged AND entry_in_son", "entry_ctid").
			Else("NULL"), "entry_ctid")).
		FromSelect(RemoveNonRelevantQ, "t1")
}

func sqTopEntryV(branchID int64, requestedCommit CommitID, lineage []lineageCommit) sq.SelectBuilder {
	lineageFilter, _, isDeletedAlias := sqLineageConditions(branchID, lineage)
	baseSelect := sq.Select().
		FromSelect(sqEntriesV(requestedCommit), "e\n").
		Where(lineageFilter).
		Columns("e.path", "e.branch_id AS source_branch",
			"e.is_committed", "e.is_tombstone").
		Column(isDeletedAlias)
	minSelect := sq.Select(" path").
		FromSelect(baseSelect, "e").
		Where("not e.is_deleted")
	return minSelect
}

func sqListByPrefix(prefix, after, delimiter string, branchID int64, maxLines int, requestedCommit CommitID, lineage []lineageCommit) sq.SelectBuilder {
	if strings.HasSuffix(after, delimiter) {
		after += DirectoryTermination
	}
	prefixLen := len(prefix) + 1
	endOfPrefixRange := prefix + DirectoryTermination
	strPosV := sq.Expr("strPos(substr(e.path,?),?)", prefixLen, delimiter)
	pathWithOutPrefixV := sq.Expr("substr(e.path,?)", prefixLen)
	directoryPartV := sq.ConcatExpr("left(", pathWithOutPrefixV, ",", strPosV, ")")
	tmp := sq.Case().When(sq.ConcatExpr(strPosV, " > 0\n"), sq.ConcatExpr(directoryPartV, " || chr(1000000) \n")).
		Else(pathWithOutPrefixV)
	getNextMarkerV := sq.ConcatExpr("\n", tmp, "\n")
	cteStart := sq.Select("1 as num").Column(sq.Alias(getNextMarkerV, "marker")).
		FromSelect(sq.Select("min(path) as path").
			FromSelect(sqTopEntryV(branchID, requestedCommit, lineage), "e").
			Where(" e.path > ? and e.path < ? ", prefix+after, endOfPrefixRange), "e")
	nextMarkerSelect := sq.Select().FromSelect(
		sq.Select("min(path) as path").FromSelect(
			sq.Select("path").FromSelect(
				sqTopEntryV(branchID, requestedCommit, lineage), "e").
				Where(" e.path > ?  || d.marker and e.path < ? ", prefix, endOfPrefixRange), "e"), "e").
		Column(getNextMarkerV)

	dirListV := sq.Select("1 as num", "marker").
		Prefix(`WITH RECURSIVE dir_list AS (`).
		FromSelect(cteStart, "t").
		SuffixExpr(
			sq.ConcatExpr("\nUNION ALL\n",
				sq.Select("d.num + 1 as num").
					Column( // calculate the next entry
						sq.ConcatExpr("(", nextMarkerSelect, ")")).
					From("dir_list as d").
					Where("num <= ? and  d.marker is not null and length(d.marker) > 0", maxLines),
				")",
				"\n SELECT marker as path",
				"\nFROM dir_list d",
				"\nWHERE d.marker IS NOT NULL"))

	return dirListV
}
