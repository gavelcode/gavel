package lcov_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/infrastructure/casefile/lcov"
)

func TestParsePerLineSingleFile(t *testing.T) {
	result, err := lcov.ParsePerLine(readFixture(t, "single_lang.lcov"))

	require.NoError(t, err)
	require.Contains(t, result, "src/foo.go")

	foo := result["src/foo.go"]
	assert.Equal(t, 1, foo[1])
	assert.Equal(t, 0, foo[2])
	assert.Equal(t, 1, foo[3])
	assert.Len(t, foo, 3)

	bar := result["src/bar.go"]
	assert.Equal(t, 1, bar[1])
	assert.Len(t, bar, 1)
}

func TestParsePerLineEmptyData(t *testing.T) {
	result, err := lcov.ParsePerLine(nil)

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestParsePerLineNoDARecords(t *testing.T) {
	data := []byte("SF:src/main.go\nLF:10\nLH:5\nend_of_record\n")

	result, err := lcov.ParsePerLine(data)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestParsePerLineMalformedDARecord(t *testing.T) {
	data := []byte("SF:src/main.go\nDA:abc,1\nend_of_record\n")

	_, err := lcov.ParsePerLine(data)

	require.Error(t, err)
	assert.ErrorIs(t, err, lcov.ErrInvalidLine)
}

func TestParsePerLineMalformedDAMissingCount(t *testing.T) {
	data := []byte("SF:src/main.go\nDA:10\nend_of_record\n")

	_, err := lcov.ParsePerLine(data)

	require.Error(t, err)
	assert.ErrorIs(t, err, lcov.ErrInvalidLine)
}

func TestParsePerLineDABeforeSFIsSkipped(t *testing.T) {
	data := []byte("DA:1,5\nSF:src/main.go\nDA:2,3\nend_of_record\n")

	result, err := lcov.ParsePerLine(data)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Contains(t, result, "src/main.go")
	assert.Len(t, result["src/main.go"], 1)
}

func TestParsePerLineMalformedDAHitCount(t *testing.T) {
	data := []byte("SF:src/main.go\nDA:1,abc\nend_of_record\n")

	_, err := lcov.ParsePerLine(data)

	require.Error(t, err)
	assert.ErrorIs(t, err, lcov.ErrInvalidLine)
}

func TestParsePerLineScannerError(t *testing.T) {
	longLine := make([]byte, 70000)
	for i := range longLine {
		longLine[i] = 'x'
	}
	data := append([]byte("SF:src/foo.go\n"), longLine...)

	_, err := lcov.ParsePerLine(data)

	require.Error(t, err)
	assert.ErrorIs(t, err, lcov.ErrScanLCOV)
}

func TestParsePerLineDAWithChecksumIsIgnored(t *testing.T) {
	data := []byte("SF:src/handler.py\nDA:1,1,xQtPgCiP5GbtaSZPWs1qaw\nDA:2,0,abc123\nend_of_record\n")

	result, err := lcov.ParsePerLine(data)

	require.NoError(t, err)
	require.Contains(t, result, "src/handler.py")
	assert.Equal(t, 1, result["src/handler.py"][1])
	assert.Equal(t, 0, result["src/handler.py"][2])
}

func TestParsePerLineMultipleFiles(t *testing.T) {
	data := []byte("SF:a.go\nDA:1,5\nend_of_record\nSF:b.go\nDA:2,0\nend_of_record\n")

	result, err := lcov.ParsePerLine(data)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 5, result["a.go"][1])
	assert.Equal(t, 0, result["b.go"][2])
}

func TestParsePerLineDuplicateFileMergesWithMax(t *testing.T) {
	data := []byte(
		"SF:shared.go\nDA:1,10\nDA:2,5\nDA:3,0\nLF:3\nLH:2\nend_of_record\n" +
			"SF:shared.go\nDA:1,0\nDA:2,8\nDA:3,0\nLF:3\nLH:1\nend_of_record\n",
	)

	result, err := lcov.ParsePerLine(data)

	require.NoError(t, err)
	require.Contains(t, result, "shared.go")
	shared := result["shared.go"]
	assert.Equal(t, 10, shared[1])
	assert.Equal(t, 8, shared[2])
	assert.Equal(t, 0, shared[3])
}
