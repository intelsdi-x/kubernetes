package discovery

import (
	"bytes"
	"io/ioutil"
	"strconv"
)

func DiscoverHugepageNumber() (int, error) {
        hugepageInfoPath := "/proc/sys/vm/nr_hugepages"
        totalBytes, err := ioutil.ReadFile(hugepageInfoPath)
        if err != nil {
                return -1, err
        }
        hpTotal := bytes.TrimSpace(totalBytes)
        hpTotalNum, err := strconv.Atoi(string(hpTotal))
        if err != nil {
                return -1, err
        }
        return hpTotalNum, nil
}