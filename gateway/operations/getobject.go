package operations

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/treeverse/lakefs/httputil"

	"github.com/treeverse/lakefs/block"
	"github.com/treeverse/lakefs/db"
	gatewayerrors "github.com/treeverse/lakefs/gateway/errors"
	ghttp "github.com/treeverse/lakefs/gateway/http"
	"github.com/treeverse/lakefs/gateway/serde"
	"github.com/treeverse/lakefs/permissions"

	"golang.org/x/xerrors"
)

type GetObject struct{}

func (controller *GetObject) RequiredPermissions(_ *http.Request, repoID, _, path string) ([]permissions.Permission, error) {
	return []permissions.Permission{
		{
			Action:   permissions.ReadObjectAction,
			Resource: permissions.ObjectArn(repoID, path),
		},
	}, nil
}

func (controller *GetObject) Handle(o *PathOperation) {
	o.Incr("get_object")
	query := o.Request.URL.Query()
	if _, exists := query["versioning"]; exists {
		o.EncodeResponse(serde.VersioningConfiguration{}, http.StatusOK)
		return
	}

	if _, exists := query["tagging"]; exists {
		o.EncodeResponse(serde.Tagging{}, http.StatusOK)
		return
	}

	beforeMeta := time.Now()
	// make sure we work on uncommitted data
	entry, err := o.Cataloger.GetEntry(o.Context(), o.Repository.Name, o.Reference, o.Path)
	metaTook := time.Since(beforeMeta)
	o.Log().
		WithField("took", metaTook).
		WithError(err).
		Debug("metadata operation to retrieve object done")

	if xerrors.Is(err, db.ErrNotFound) {
		// TODO: create distinction between missing repo & missing key
		o.EncodeError(gatewayerrors.Codes.ToAPIErr(gatewayerrors.ErrNoSuchKey))
		return
	}
	if err != nil {
		o.EncodeError(gatewayerrors.Codes.ToAPIErr(gatewayerrors.ErrInternalError))
		return
	}

	o.SetHeader("Last-Modified", httputil.HeaderTimestamp(entry.CreationDate))
	o.SetHeader("ETag", httputil.ETag(entry.Checksum))
	o.SetHeader("Accept-Ranges", "bytes")
	// TODO: the rest of https://docs.aws.amazon.com/en_pv/AmazonS3/latest/API/API_GetObject.html

	// now we might need the object itself
	ent, err := o.Cataloger.GetEntry(o.Context(), o.Repository.Name, o.Reference, o.Path)
	if err != nil {
		o.EncodeError(gatewayerrors.Codes.ToAPIErr(gatewayerrors.ErrInternalError))
		return
	}

	// range query
	var expected int64
	var data io.ReadCloser
	var rng ghttp.HttpRange
	rng.StartOffset = -1
	// range query
	rangeSpec := o.Request.Header.Get("Range")
	if len(rangeSpec) > 0 {
		rng, err = ghttp.ParseHTTPRange(rangeSpec, ent.Size)
		if err != nil {
			o.Log().WithError(err).WithField("range", rangeSpec).Debug("invalid range spec")
		}
	}
	if rangeSpec == "" || err != nil {
		// assemble a response body (range-less query)
		expected = ent.Size
		data, err = o.BlockStore.Get(block.ObjectPointer{StorageNamespace: o.Repository.StorageNamespace, Identifier: ent.PhysicalAddress})
	} else {
		expected = rng.EndOffset - rng.StartOffset + 1 // both range ends are inclusive
		data, err = o.BlockStore.GetRange(block.ObjectPointer{StorageNamespace: o.Repository.StorageNamespace, Identifier: ent.PhysicalAddress}, rng.StartOffset, rng.EndOffset)
	}
	if err != nil {
		o.EncodeError(gatewayerrors.Codes.ToAPIErr(gatewayerrors.ErrInternalError))
		return
	}
	defer func() {
		_ = data.Close()
	}()
	o.SetHeader("Content-Length", fmt.Sprintf("%d", expected))
	if rng.StartOffset != -1 {
		o.SetHeader("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rng.StartOffset, rng.EndOffset, ent.Size))
	}
	_, err = io.Copy(o.ResponseWriter, data)
	if err != nil {
		o.Log().WithError(err).Error("could not write response body for object")
	}
}
