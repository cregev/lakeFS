// Code generated by go-swagger; DO NOT EDIT.

package objects

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/treeverse/lakefs/api/gen/models"
)

// GetObjectReader is a Reader for the GetObject structure.
type GetObjectReader struct {
	formats strfmt.Registry
	writer  io.Writer
}

// ReadResponse reads a server response into the received o.
func (o *GetObjectReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetObjectOK(o.writer)
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetObjectUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetObjectNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewGetObjectDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetObjectOK creates a GetObjectOK with default headers values
func NewGetObjectOK(writer io.Writer) *GetObjectOK {
	return &GetObjectOK{
		Payload: writer,
	}
}

/*GetObjectOK handles this case with default header values.

object content
*/
type GetObjectOK struct {
	ContentLength int64

	ETag string

	LastModified string

	Payload io.Writer
}

func (o *GetObjectOK) Error() string {
	return fmt.Sprintf("[GET /repositories/{repositoryId}/branches/{branchId}/objects][%d] getObjectOK  %+v", 200, o.Payload)
}

func (o *GetObjectOK) GetPayload() io.Writer {
	return o.Payload
}

func (o *GetObjectOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header Content-Length
	contentLength, err := swag.ConvertInt64(response.GetHeader("Content-Length"))
	if err != nil {
		return errors.InvalidType("Content-Length", "header", "int64", response.GetHeader("Content-Length"))
	}
	o.ContentLength = contentLength

	// response header ETag
	o.ETag = response.GetHeader("ETag")

	// response header Last-Modified
	o.LastModified = response.GetHeader("Last-Modified")

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetObjectUnauthorized creates a GetObjectUnauthorized with default headers values
func NewGetObjectUnauthorized() *GetObjectUnauthorized {
	return &GetObjectUnauthorized{}
}

/*GetObjectUnauthorized handles this case with default header values.

Unauthorized
*/
type GetObjectUnauthorized struct {
	Payload *models.Error
}

func (o *GetObjectUnauthorized) Error() string {
	return fmt.Sprintf("[GET /repositories/{repositoryId}/branches/{branchId}/objects][%d] getObjectUnauthorized  %+v", 401, o.Payload)
}

func (o *GetObjectUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetObjectUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetObjectNotFound creates a GetObjectNotFound with default headers values
func NewGetObjectNotFound() *GetObjectNotFound {
	return &GetObjectNotFound{}
}

/*GetObjectNotFound handles this case with default header values.

path or branch not found
*/
type GetObjectNotFound struct {
	Payload *models.Error
}

func (o *GetObjectNotFound) Error() string {
	return fmt.Sprintf("[GET /repositories/{repositoryId}/branches/{branchId}/objects][%d] getObjectNotFound  %+v", 404, o.Payload)
}

func (o *GetObjectNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetObjectNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetObjectDefault creates a GetObjectDefault with default headers values
func NewGetObjectDefault(code int) *GetObjectDefault {
	return &GetObjectDefault{
		_statusCode: code,
	}
}

/*GetObjectDefault handles this case with default header values.

generic error response
*/
type GetObjectDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get object default response
func (o *GetObjectDefault) Code() int {
	return o._statusCode
}

func (o *GetObjectDefault) Error() string {
	return fmt.Sprintf("[GET /repositories/{repositoryId}/branches/{branchId}/objects][%d] getObject default  %+v", o._statusCode, o.Payload)
}

func (o *GetObjectDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetObjectDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
