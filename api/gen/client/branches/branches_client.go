// Code generated by go-swagger; DO NOT EDIT.

package branches

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new branches API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for branches API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientService is the interface for Client methods
type ClientService interface {
	CreateBranch(params *CreateBranchParams, authInfo runtime.ClientAuthInfoWriter) (*CreateBranchCreated, error)

	DeleteBranch(params *DeleteBranchParams, authInfo runtime.ClientAuthInfoWriter) (*DeleteBranchNoContent, error)

	DiffBranch(params *DiffBranchParams, authInfo runtime.ClientAuthInfoWriter) (*DiffBranchOK, error)

	GetBranch(params *GetBranchParams, authInfo runtime.ClientAuthInfoWriter) (*GetBranchOK, error)

	ListBranches(params *ListBranchesParams, authInfo runtime.ClientAuthInfoWriter) (*ListBranchesOK, error)

	RevertBranch(params *RevertBranchParams, authInfo runtime.ClientAuthInfoWriter) (*RevertBranchNoContent, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  CreateBranch creates branch
*/
func (a *Client) CreateBranch(params *CreateBranchParams, authInfo runtime.ClientAuthInfoWriter) (*CreateBranchCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateBranchParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createBranch",
		Method:             "POST",
		PathPattern:        "/repositories/{repositoryId}/branches",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreateBranchReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateBranchCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreateBranchDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  DeleteBranch deletes branch
*/
func (a *Client) DeleteBranch(params *DeleteBranchParams, authInfo runtime.ClientAuthInfoWriter) (*DeleteBranchNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteBranchParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "deleteBranch",
		Method:             "DELETE",
		PathPattern:        "/repositories/{repositoryId}/branches/{branchId}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteBranchReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteBranchNoContent)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteBranchDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  DiffBranch diffs branch
*/
func (a *Client) DiffBranch(params *DiffBranchParams, authInfo runtime.ClientAuthInfoWriter) (*DiffBranchOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDiffBranchParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "diffBranch",
		Method:             "GET",
		PathPattern:        "/repositories/{repositoryId}/branches/{branchId}/diff",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DiffBranchReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DiffBranchOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DiffBranchDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetBranch gets branch
*/
func (a *Client) GetBranch(params *GetBranchParams, authInfo runtime.ClientAuthInfoWriter) (*GetBranchOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetBranchParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getBranch",
		Method:             "GET",
		PathPattern:        "/repositories/{repositoryId}/branches/{branchId}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetBranchReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetBranchOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetBranchDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  ListBranches lists branches
*/
func (a *Client) ListBranches(params *ListBranchesParams, authInfo runtime.ClientAuthInfoWriter) (*ListBranchesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListBranchesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "listBranches",
		Method:             "GET",
		PathPattern:        "/repositories/{repositoryId}/branches",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ListBranchesReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListBranchesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*ListBranchesDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  RevertBranch reverts branch to specified commit or revert specific path changes to last commit pipe if nothing passed reverts all non committed changes
*/
func (a *Client) RevertBranch(params *RevertBranchParams, authInfo runtime.ClientAuthInfoWriter) (*RevertBranchNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewRevertBranchParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "revertBranch",
		Method:             "PUT",
		PathPattern:        "/repositories/{repositoryId}/branches/{branchId}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &RevertBranchReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*RevertBranchNoContent)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*RevertBranchDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
