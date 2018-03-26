package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	CHROME     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36"
	FIREFOX    = "Mozilla/5.0 (Windows NT 5.1; rv:7.0.1) Gecko/20100101 Firefox/7.0.1"
	IE         = "Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1)"
	SAFARI     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7"
	ANDROID    = "Mozilla/5.0 (Linux; U; Android 4.0.4; pt-br; MZ608 Build/7.7.1-141-7-FLEM-UMTS-LA) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Safari/534.30"
	OPERA      = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36 OPR/43.0.2442.991"
	OPERA_MINI = "Opera/9.80 (J2ME/MIDP; Opera Mini/4.2/28.3590; U; en) Presto/2.8.119 Version/11.10"

	GIT_LFS = "git-lfs/2.0.1 (GitHub; windows amd64; go 1.8; git 678cdbd4)"
)

func TestIsBrowserRequest(t *testing.T) {
	assert.True(t, IsBrowserRequest(CHROME))
	assert.True(t, IsBrowserRequest(FIREFOX))
	assert.True(t, IsBrowserRequest(IE))
	assert.True(t, IsBrowserRequest(SAFARI))
	assert.True(t, IsBrowserRequest(ANDROID))
	assert.True(t, IsBrowserRequest(OPERA))
	assert.True(t, IsBrowserRequest(OPERA_MINI))
	assert.False(t, IsBrowserRequest(GIT_LFS))
}
