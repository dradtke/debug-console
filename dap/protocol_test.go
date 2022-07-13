package dap_test

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/dradtke/debug-console/dap"
)

func TestReadMessage(t *testing.T) {
	var (
		testBody    = "hello world"
		testHeaders = fmt.Sprintf("Content-Length: %d\r\n", len(testBody))
		r           = strings.NewReader(fmt.Sprintf("%s\r\n%s", testHeaders, testBody))
		scratch     = make([]byte, 4096)
		buf         = bytes.Buffer{}
	)

	headers, _, body, err := dap.ReadMessage(r, scratch, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(headers) != 1 {
		t.Errorf("unexpected number of headers: %d", len(headers))
	}
	if contentLength := headers["Content-Length"]; contentLength != strconv.Itoa(len(testBody)) {
		t.Errorf("unexpected content length: %s != %s", contentLength, strconv.Itoa(len(testBody)))
	}

	if body != testBody {
		t.Errorf("unexpected body: %s != %s", body, testBody)
	}

	if buf.Len() != 0 {
		t.Errorf("unexpected data left over in buffer: %s", buf.String())
	}
}
