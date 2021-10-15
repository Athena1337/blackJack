package runner

import (
	"blackJack/utils"
	"io"
	"io/ioutil"
)

func (r *Runner) GetFaviconHash(url string) (hash string, err error) {
	request, err := newRequest("GET", url)
	if err != nil {
		return
	}
	resp, err := r.client.Do(request)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body) //nolint

	hash = utils.Mmh3Hash32(utils.StandBase64(body))
	return
}
