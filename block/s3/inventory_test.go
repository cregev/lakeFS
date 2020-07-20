package s3_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"testing"

	s32 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/treeverse/lakefs/block/s3"
)

func rows(keys ...string) []s3.ParquetInventoryObject {
	if keys == nil {
		return nil
	}
	res := make([]s3.ParquetInventoryObject, len(keys))
	latest := true
	for i, key := range keys {
		res[i].Key = key
		res[i].IsLatest = &latest
	}
	return res
}

var fileContents = map[string][]string{
	"f1": {"f1row1", "f1row2"},
	"f2": {"f2row1", "f2row2"},
	"f3": {"f3row1", "f3row2"},
	"f4": {"f4row1", "f4row2", "f4row3", "f4row4", "f4row5", "f4row6", "f4row7"},
	"f5": {"a1", "a2", "a3"},
	"f6": {"a4", "a5", "a6", "a7"},
}

func TestFetch(t *testing.T) {
	testdata := []struct {
		InventoryFiles  []string
		ExpectedObjects []string
		ReadBatchSize   int
		ErrIndex        int
	}{
		{
			InventoryFiles:  []string{"f1", "f2", "f3"},
			ExpectedObjects: []string{"f1row1", "f1row2", "f2row1", "f2row2", "f3row1", "f3row2"},
		},
		{
			InventoryFiles:  []string{"f1", "f2", "f3"},
			ExpectedObjects: []string{"f1row1", "f1row2", "f2row1", "f2row2", "f3row1", "f3row2"},
			ReadBatchSize:   10000,
		},
		{
			InventoryFiles:  []string{"f1", "f2", "f3"},
			ExpectedObjects: []string{"f1row1", "f1row2", "f2row1", "f2row2", "f3row1", "f3row2"},
			ReadBatchSize:   3,
		},
		{
			InventoryFiles:  []string{"f1", "f2", "f3"},
			ExpectedObjects: []string{"f1row1", "f1row2", "f2row1", "f2row2", "f3row1", "f3row2"},
			ReadBatchSize:   6,
		},
		{
			InventoryFiles:  []string{},
			ExpectedObjects: []string{},
		},
		{
			InventoryFiles:  []string{"f4"},
			ExpectedObjects: []string{"f4row1", "f4row2", "f4row3", "f4row4", "f4row5", "f4row6", "f4row7"},
		},
		{
			InventoryFiles:  []string{"f1", "f4"},
			ExpectedObjects: []string{"f1row1", "f1row2", "f4row1", "f4row2", "f4row3", "f4row4", "f4row5", "f4row6", "f4row7"},
		},
		{
			InventoryFiles:  []string{"f5", "f6"},
			ExpectedObjects: []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7"},
		},
		{
			InventoryFiles:  []string{"f5", "f6"},
			ExpectedObjects: []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7"},
			ReadBatchSize:   2,
		},
	}

	manifestURL := "s3://example-bucket/manifest1.json"
	for _, test := range testdata {
		inv, err := s3.GenerateInventory(manifestURL, &mockS3Client{
			FilesByManifestURL: map[string][]string{manifestURL: test.InventoryFiles},
		})
		s3Inv := inv.(*s3.Inventory)
		s3Inv.GetParquetReader = mockParquetReaderGetter
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		it, err := inv.Iterator(context.Background())
		it.(*s3.InventoryIterator).ReadBatchSize = test.ReadBatchSize
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		objects := make([]string, 0, len(test.ExpectedObjects))
		for it.Next() {
			objects = append(objects, it.Get().Key)
		}
		if len(objects) != len(test.ExpectedObjects) {
			t.Fatalf("unexpected number of objects in inventory. expected=%d, got=%d", len(test.ExpectedObjects), len(objects))
		}
		if !reflect.DeepEqual(objects, test.ExpectedObjects) {
			t.Fatalf("objects in inventory differrent than expected. expected=%v, got=%v", test.ExpectedObjects, objects)
		}
	}
}

type mockParquetReader struct {
	rows     []s3.ParquetInventoryObject
	nextIdx  int
	errIndex *int
}

func (m *mockParquetReader) Read(dstInterface interface{}) error {
	res := make([]s3.ParquetInventoryObject, 0, len(m.rows))
	dst := dstInterface.(*[]s3.ParquetInventoryObject)
	for i := m.nextIdx; i < len(m.rows) && i < m.nextIdx+len(*dst); i++ {
		if m.errIndex != nil && i == *m.errIndex {
			return errors.New("mock parquet reader reached error index")
		}
		res = append(res, m.rows[i])
	}
	m.nextIdx = m.nextIdx + len(res)
	*dst = res
	return nil
}

func (m *mockParquetReader) GetNumRows() int64 {
	return int64(len(m.rows))
}

func mockParquetReaderGetter(ctx context.Context, svc s3iface.S3API, bucket string, key string) (s3.ParquetReader, error) {
	if bucket != "example-bucket" {
		return nil, fmt.Errorf("wrong bucket name: %s", bucket)
	}
	return &mockParquetReader{rows: rows(fileContents[key]...)}, nil
}

func (m *mockS3Client) GetObject(input *s32.GetObjectInput) (*s32.GetObjectOutput, error) {
	output := s32.GetObjectOutput{}
	manifestURL := fmt.Sprintf("s3://%s%s", *input.Bucket, *input.Key)
	if !manifestExists(manifestURL) {
		return &output, nil
	}
	manifestFileNames := m.FilesByManifestURL[manifestURL]
	if manifestFileNames == nil {
		manifestFileNames = []string{"inventory/lakefs-example-data/my_inventory/data/ea8268b2-a6ba-42de-8694-91a9833b4ff1.parquet"}
	}
	manifestFiles := make([]interface{}, len(manifestFileNames))
	for _, filename := range manifestFileNames {
		manifestFiles = append(manifestFiles, struct {
			Key string `json:"key"`
		}{
			Key: filename,
		})
	}
	filesJSON, err := json.Marshal(manifestFiles)
	if err != nil {
		return nil, err
	}
	destBucket := m.DestBucket
	if m.DestBucket == "" {
		destBucket = "example-bucket"
	}
	reader := strings.NewReader(fmt.Sprintf(`{
  "sourceBucket" : "lakefs-example-data",
  "destinationBucket" : "arn:aws:s3:::%s",
  "version" : "2016-11-30",
  "creationTimestamp" : "1593216000000",
  "fileFormat" : "Parquet",
  "fileSchema" : "message s3.inventory {  required binary bucket (STRING);  required binary key (STRING);  optional binary version_id (STRING);  optional boolean is_latest;  optional boolean is_delete_marker;  optional int64 size;  optional int64 last_modified_date (TIMESTAMP(MILLIS,true));  optional binary e_tag (STRING);  optional binary storage_class (STRING);  optional boolean is_multipart_uploaded;}",
  "files" : %s}`, destBucket, filesJSON))
	return output.SetBody(ioutil.NopCloser(reader)), nil
}

type mockS3Client struct {
	s3iface.S3API
	FilesByManifestURL map[string][]string
	DestBucket         string
}

func manifestExists(manifestURL string) bool {
	match, _ := regexp.MatchString("s3://example-bucket/manifest[0-9]+.json", manifestURL)
	return match
}
