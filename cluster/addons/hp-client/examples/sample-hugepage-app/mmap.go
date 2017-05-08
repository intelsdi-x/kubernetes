package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"strconv"
	"github.com/golang/glog"
)
var mmap []byte

//Hardcoded constants for Demo
const twoMBinBytes = 2097152
const mountPath = "/mnt/huge/"

func logsHugepagesInfo() (err error) {
	//Open /proc/meminfo file
	memInfoFile, err := os.Open("/proc/meminfo")
	if err != nil {
		glog.Fatalf("Error opening file: %v", err)
		return err
	}
	//Create a scanner for the /proc/meminfo file
	scanner := bufio.NewScanner(memInfoFile)
	if err != nil {
		glog.Fatalf("Error scanning the file file: %v", err)
		return err
	}
	//Start scanning the /proc/meminfo file line by line
	for scanner.Scan() {
		file_line := scanner.Text()

		if strings.Contains(file_line, "HugePages") {
			fmt.Println(file_line)
		}
	}
	memInfoFile.Close()
	return
}

func allocateHugepage(hugepageMountPath string, cgroupPath string, hugepagelength int) error {
	hp_dir := fmt.Sprintf("%s%s", hugepageMountPath, cgroupPath)
	os.MkdirAll(hp_dir, 0777)
	path_to_file := fmt.Sprintf("%s%s%s", hugepageMountPath, cgroupPath, "/pod_hp")
	_, err := os.Create(path_to_file)
	if err != nil {
		return fmt.Errorf("Error creating mmap file %v", err)
	}
	map_file, err := syscall.Open(path_to_file, os.O_RDWR, 0777)
	if err != nil {
		return fmt.Errorf("Error opening the mmap file")
	}
	mmap, err = syscall.Mmap(map_file, 0, hugepagelength, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("Error in mmap syscall %v", err)

	}
	fmt.Println("mmap syscall called successfully")
	// Use the mmap otherwise goes hugepage reserved
	//fmt.Println("mmap printing %v",mmap[0])
	hpValStr := "1"
        if len(os.Args) == 2 {
                hpValStr = os.Args[1]
        }
	tmpFilepath := fmt.Sprintf("/tmp/mmap_file%s",hpValStr)
	err = ioutil.WriteFile(tmpFilepath, mmap, 0777)
	if err != nil {
		return fmt.Errorf("Error in writing  mmap to a file /tmp/mmap", err)
	}
	//Close the mmap file
	syscall.Close(map_file)
	return nil
}

func freeHugepage(hugepageMountPath string, cgroupPath string) error {
	path_to_file := fmt.Sprintf("%s%s%s", hugepageMountPath, cgroupPath, "/pod_hp")
	err := os.Remove(path_to_file)
	if err != nil {
		return fmt.Errorf("Error deleting file %v", err)
	}
	pod_path_to_delete := fmt.Sprintf("%s%s", hugepageMountPath, cgroupPath)
	err = syscall.Rmdir(pod_path_to_delete)
	if err != nil {
		return fmt.Errorf("Cannot delete pod directory %v", err)
	}
	return nil
}
func main(){
	//By default 1 hugepage is allocated
	hpValStr := "1"
	if len(os.Args) == 2 {
		hpValStr = os.Args[1]
	}
	hpValInt, _ := strconv.Atoi(hpValStr)
	
	err := allocateHugepage(mountPath, os.Getenv("HOSTNAME"),hpValInt*twoMBinBytes)
	if err != nil {
                fmt.Println("Error in allocHugepage: %v", err)
        }

	err = logsHugepagesInfo()
	if err != nil {
		fmt.Println("Error in getting logs: %v", err)
	}
	
}
