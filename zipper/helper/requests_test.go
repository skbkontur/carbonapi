package helper

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/zipper/types"
	"github.com/stretchr/testify/assert"
)

func TestMergeHttpErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []merry.Error
		wantCode int
		want     []string
	}{
		{
			name:     "NotFound",
			errors:   []merry.Error{},
			wantCode: http.StatusNotFound,
			want:     []string{},
		},
		{
			name: "NetErr",
			errors: []merry.Error{
				types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")}).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"connect: refused"},
		},
		{
			name: "NetErr (incapsulated)",
			errors: []merry.Error{
				types.ErrMaxTriesExceeded.WithCause(types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")})).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"connect: refused"},
		},
		{
			name: "ServiceUnavailable",
			errors: []merry.Error{
				merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"unavaliable"},
		},
		{
			name: "GatewayTimeout and ServiceUnavailable",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"timeout", "unavaliable"},
		},
		{
			name: "ServiceUnavailable and GatewayTimeout",
			errors: []merry.Error{
				merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"unavaliable", "timeout"},
		},
		{
			name: "Forbidden and GatewayTimeout",
			errors: []merry.Error{
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"limit", "timeout"},
		},
		{
			name: "GatewayTimeout and Forbidden",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"timeout", "limit"},
		},
		{
			name: "InternalServerError and Forbidden",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(http.StatusInternalServerError),
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"error", "limit"},
		},
		{
			name: "InternalServerError and GatewayTimeout",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(http.StatusInternalServerError),
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"error", "timeout"},
		},
		{
			name: "GatewayTimeout and InternalServerError",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				merry.New("error").WithHTTPCode(http.StatusInternalServerError),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"timeout", "error"},
		},
		{
			name: "BadRequest and Forbidden",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(http.StatusBadRequest),
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusBadRequest,
			want:     []string{"error", "limit"},
		},
		{
			name: "Forbidden and BadRequest",
			errors: []merry.Error{
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
				merry.New("error").WithHTTPCode(http.StatusBadRequest),
			},
			wantCode: http.StatusBadRequest,
			want:     []string{"limit", "error"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, got := MergeHttpErrors(tt.errors)
			if gotCode != tt.wantCode {
				t.Errorf("MergeHttpErrors() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeHttpErrors() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeHttpErrorMap(t *testing.T) {
	tests := []struct {
		name     string
		errors   map[string]merry.Error
		wantCode int
		want     []string
	}{
		{
			name:     "NotFound",
			errors:   map[string]merry.Error{},
			wantCode: http.StatusNotFound,
			want:     []string{},
		},
		{
			name: "NetErr",
			errors: map[string]merry.Error{
				"a": types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")}).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"a: connect: refused"},
		},
		{
			name: "NetErr (incapsulated)",
			errors: map[string]merry.Error{
				"b": types.ErrMaxTriesExceeded.WithCause(types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")})).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"b: connect: refused"},
		},
		{
			name: "ServiceUnavailable",
			errors: map[string]merry.Error{
				"d": merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"d: unavaliable"},
		},
		{
			name: "GatewayTimeout and ServiceUnavailable",
			errors: map[string]merry.Error{
				"a":  merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				"de": merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"a: timeout", "de: unavaliable"},
		},
		{
			name: "ServiceUnavailable and GatewayTimeout",
			errors: map[string]merry.Error{
				"de": merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
				"a":  merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"a: timeout", "de: unavaliable"},
		},
		{
			name: "Forbidden and GatewayTimeout",
			errors: map[string]merry.Error{
				"de": merry.New("limit").WithHTTPCode(http.StatusForbidden),
				"c":  merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"c: timeout", "de: limit"},
		},
		{
			name: "GatewayTimeout and Forbidden",
			errors: map[string]merry.Error{
				"a": merry.New("limit").WithHTTPCode(http.StatusForbidden),
				"c": merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"a: limit", "c: timeout"},
		},
		{
			name: "InternalServerError and Forbidden",
			errors: map[string]merry.Error{
				"a":  merry.New("error").WithHTTPCode(http.StatusInternalServerError),
				"cd": merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"a: error", "cd: limit"},
		},
		{
			name: "InternalServerError and GatewayTimeout",
			errors: map[string]merry.Error{
				"a": merry.New("error").WithHTTPCode(http.StatusInternalServerError),
				"b": merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"a: error", "b: timeout"},
		},
		{
			name: "GatewayTimeout and InternalServerError",
			errors: map[string]merry.Error{
				"a":  merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				"cd": merry.New("error").WithHTTPCode(http.StatusInternalServerError),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"a: timeout", "cd: error"},
		},
		{
			name: "BadRequest and Forbidden",
			errors: map[string]merry.Error{
				"de": merry.New("error").WithHTTPCode(http.StatusBadRequest),
				"a":  merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusBadRequest,
			want:     []string{"a: limit", "de: error"},
		},
		{
			name: "Forbidden and BadRequest",
			errors: map[string]merry.Error{
				"a": merry.New("limit").WithHTTPCode(http.StatusForbidden),
				"b{c,de,klmn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.b{c,de,klmn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}": merry.New("error").WithHTTPCode(http.StatusBadRequest),
			},
			wantCode: http.StatusBadRequest,
			want: []string{
				"a: limit",
				"b{c,de,klmn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.b{c,de,k...mn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}: error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, got := MergeHttpErrorMap(tt.errors)
			if gotCode != tt.wantCode {
				t.Errorf("MergeHttpErrors() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
			// sort error strings
			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeHttpErrors() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripHtmlTags(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{
			name:   "Empty",
			s:      "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "Broken #1",
			s:      "<html>\r\n<head",
			maxLen: 0,
			want:   "head",
		},
		{
			name:   "Broken #2",
			s:      "<html>\r\nhead>",
			maxLen: 0,
			want:   "head>",
		},
		{
			name:   "Nginx Gateway Timeout",
			s:      "<html>\r\n<head><title>504 Gateway Time-out</title></head>\r\n<body>\r\n<center><h1>504 Gateway Time-out</h1></center>\r\n<hr><center>nginx</center>\r\n</body>\r\n</html>\r",
			maxLen: 0,
			want:   "504 Gateway Time-out\r\n\r\n504 Gateway Time-out\r\nnginx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stripHtmlTags(tt.s, tt.maxLen))
		})
	}
}

func Test_recalcCode(t *testing.T) {
	tests := []struct {
		code    int
		newCode int
		want    int
	}{
		{code: 500, newCode: 403, want: 403},
		{code: 403, newCode: 500, want: 403},
		{code: 403, newCode: 400, want: 400},
		{code: 400, newCode: 403, want: 400},
		{code: 500, newCode: 503, want: 500},
		{code: 503, newCode: 500, want: 500},
		{code: 503, newCode: 502, want: 503},
		{code: 0, newCode: 502, want: 503},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := recalcCode(tt.code, tt.newCode); got != tt.want {
				t.Errorf("recalcCode(%d, %d) = %d, want %d", tt.code, tt.newCode, got, tt.want)
			}
		})
	}
}
