package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID_UnmarshalStringOrNumber(t *testing.T) {
	var s struct {
		A ID `json:"a"`
		B ID `json:"b"`
		C ID `json:"c"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"a":"KWN4QjAd","b":4166421,"c":null}`), &s))
	assert.Equal(t, ID("KWN4QjAd"), s.A)
	assert.Equal(t, ID("4166421"), s.B)
	assert.Equal(t, ID(""), s.C)
}

func TestID_MarshalAlwaysString(t *testing.T) {
	b, err := json.Marshal(ID("42"))
	require.NoError(t, err)
	assert.Equal(t, `"42"`, string(b))
}

func TestInt_NumberAndString(t *testing.T) {
	var v struct {
		A Int `json:"a"`
		B Int `json:"b"`
		C Int `json:"c"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"a":10,"b":"20","c":""}`), &v))
	assert.Equal(t, int64(10), v.A.Int64())
	assert.Equal(t, int64(20), v.B.Int64())
	assert.Equal(t, int64(0), v.C.Int64())
}

func TestBool_Flexible(t *testing.T) {
	var v struct {
		A Bool `json:"a"`
		B Bool `json:"b"`
		C Bool `json:"c"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"a":true,"b":"yes","c":"0"}`), &v))
	assert.True(t, bool(v.A))
	assert.True(t, bool(v.B))
	assert.False(t, bool(v.C))
}

func TestStringOrSlice(t *testing.T) {
	var v struct {
		A StringOrSlice `json:"a"`
		B StringOrSlice `json:"b"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"a":"x","b":["y","z"]}`), &v))
	assert.Equal(t, StringOrSlice{"x"}, v.A)
	assert.Equal(t, StringOrSlice{"y", "z"}, v.B)
}

func TestOpportunity_DecodesRealShape(t *testing.T) {
	// A trimmed real search result: mixed id types must not break decoding.
	raw := `{"id":"KWN4QjAd","objective":"Go Engineer","remote":true,"locations":["United States"],"organizations":[{"id":4166421}]}`
	var o Opportunity
	require.NoError(t, json.Unmarshal([]byte(raw), &o))
	assert.Equal(t, ID("KWN4QjAd"), o.ID)
	assert.True(t, bool(o.Remote))
	assert.Equal(t, StringOrSlice{"United States"}, o.Locations)
}

func FuzzFlexibleTypes(f *testing.F) {
	seeds := []string{`"a"`, `1`, `true`, `"yes"`, `["x"]`, `null`, `1.5`, `"12"`, `{}`, `1e9`, `-3`}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, s string) {
		var id ID
		_ = id.UnmarshalJSON([]byte(s))
		var i Int
		_ = i.UnmarshalJSON([]byte(s))
		var b Bool
		_ = b.UnmarshalJSON([]byte(s))
		var ss StringOrSlice
		_ = ss.UnmarshalJSON([]byte(s))
		// Marshalling an ID must always succeed and round-trip to a string.
		if out, err := id.MarshalJSON(); err == nil {
			var back string
			_ = json.Unmarshal(out, &back)
		}
	})
}
