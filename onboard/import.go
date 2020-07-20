package onboard

import (
	"context"
	"fmt"
	"github.com/treeverse/lakefs/block"
	"github.com/treeverse/lakefs/catalog"
	"time"
)

const (
	DefaultBranchName = "import-from-inventory"
	CommitMsgTemplate = "Import from %s"
)

type Importer struct {
	repository         string
	batchSize          int
	inventoryGenerator block.InventoryGenerator
	inventory          block.Inventory
	CatalogActions     RepoActions
}

type InventoryImportStats struct {
	AddedOrChanged       int
	Deleted              int
	DryRun               bool
	PreviousInventoryURL string
	PreviousImportDate   time.Time
}

type ObjectImport struct {
	Obj      block.InventoryObject
	ToDelete bool
}

type InventoryImport struct {
	addIterator    block.InventoryIterator
	deleteIterator block.InventoryIterator
	errChannel     <-chan error
	stats          InventoryImportStats
}

func CreateImporter(cataloger catalog.Cataloger, inventoryGenerator block.InventoryGenerator, username string, inventoryURL string, repository string) (importer *Importer, err error) {
	res := &Importer{
		repository:         repository,
		inventoryGenerator: inventoryGenerator,
	}
	res.inventory, err = inventoryGenerator.GenerateInventory(inventoryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}
	res.CatalogActions = NewCatalogActions(cataloger, repository, username)
	return res, nil
}

func (s *Importer) diffFromCommit(ctx context.Context, commit catalog.CommitLog) (*DiffIterator, error) {
	previousInventoryURL := ExtractInventoryURL(commit.Metadata)
	if previousInventoryURL == "" {
		return nil, fmt.Errorf("no inventory_url in commit Metadata. commit_ref=%s", commit.Reference)
	}
	previousInv, err := s.inventoryGenerator.GenerateInventory(previousInventoryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory for previous state: %w", err)
	}
	previousObjs, err := previousInv.Iterator(ctx)
	if err != nil {
		return nil, err
	}
	currentObjs, err := s.inventory.Iterator(ctx)
	if err != nil {
		return nil, err
	}
	return NewDiffIterator(previousObjs, currentObjs), nil
}

func (s *Importer) Import(ctx context.Context, dryRun bool) (*InventoryImportStats, error) {
	previousCommit, err := s.CatalogActions.GetPreviousCommit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous commit: %w", err)
	}
	var dataToImport *DiffIterator
	if previousCommit == nil {
		// no previous commit, add whole inventory
		it, err := s.inventory.Iterator(ctx)
		if err != nil {
			return nil, err
		}
		dataToImport = NewDiffIterator(nil, it)
	} else {
		dataToImport, err = s.diffFromCommit(ctx, *previousCommit)
		if err != nil {
			return nil, err
		}
	}
	stats, err := s.CatalogActions.CreateAndDeleteObjects(ctx, *dataToImport, dryRun)
	if err != nil {
		return nil, err
	}
	stats.DryRun = dryRun
	if previousCommit != nil {
		stats.PreviousImportDate = previousCommit.CreationDate
		stats.PreviousInventoryURL = previousCommit.Metadata["inventory_url"]
	}
	if !dryRun {
		commitMetadata := CreateCommitMetadata(s.inventory, *stats)
		err = s.CatalogActions.Commit(ctx, fmt.Sprintf(CommitMsgTemplate, s.inventory.SourceName()), commitMetadata)
		if err != nil {
			return nil, err
		}
	}
	return stats, nil
}
