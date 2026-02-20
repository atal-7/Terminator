package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const max_iterations int = 8

func main() {

	for i := 0; i <= max_iterations; {
		reader := bufio.NewReader(os.Stdin)
		input := get_selection(reader)
		if input == 0 {
			fmt.Println("terminating sdc main, dc and stryker api")
			kill_sdc_apps()
		} else {
			fmt.Println("Please enter the name of the process")
			app_name_raw, read_err := reader.ReadString('\n')
			if read_err != nil {
				//log
				fmt.Println("Error when reading")
				return
			}

			app_name := strings.TrimSpace(app_name_raw)
			overall_result, err := kill_process_by_name(app_name) // revist
			if err != nil {
				fmt.Println("Error when killing process")
				return
			}
			compute_result(overall_result, app_name)
		}

		// try again
		fmt.Println("\nEnter y to continue. Anything else to exit")
		input_raw, read_err := reader.ReadString('\n')
		if read_err != nil {
			//log
			fmt.Println("Error when reading")
			return
		}
		if try_again_str := strings.TrimSpace(input_raw); !strings.EqualFold(try_again_str, "y") {
			return
		}
	}
}

func compute_result(overall_result []bool, app_name string) {
	success := 0
	failure := 0
	for _, result := range overall_result {
		if result {
			success++
		} else {
			failure++
		}
	}

	if failure > 0 {
		fmt.Println("Process Report for :", app_name)
		fmt.Println("Total deleted ", success)
		fmt.Println("Total failed ", failure)
	} else if success > 0 {
		fmt.Println("Deleted ", success, " instance of ", app_name, " with success")
	} else {
		fmt.Println("Nothing to delete for ", app_name)
	}
}

func get_selection(reader *bufio.Reader) int {

	result := -1

	for i := 0; i <= max_iterations; {
		fmt.Println("Make a selection")
		fmt.Println("1: to kill a particular app")
		fmt.Println("0: to kill sdc (main, dc and strykerapi)")

		input_raw, read_err := reader.ReadString('\n')
		if read_err != nil {
			//log
			fmt.Println("Error when reading")
			return -1
		}
		input := strings.TrimSpace(input_raw)
		val, err := strconv.Atoi(input)
		if err != nil || (val != 0 && val != 1) {
			fmt.Println("Oi mate select a valid option")
			fmt.Println("Try again")
		} else {
			result = val
			break
		}
	}

	return result
}

func kill_sdc_apps() error {
	const sdc_main string = "sdcmain.exe"
	const device_control string = "sdcdevicecontrolapplication.exe"
	const stryker_api string = "strykerapiserver.exe"
	const sdc_stryker_api string = "sdcstrykerapiserver.exe"

	overall_result, err := kill_process_by_name(device_control)
	if err != nil {
		fmt.Println("Error when killing process")
		return err
	}
	compute_result(overall_result, device_control)

	overall_result, err = kill_process_by_name(sdc_main)
	if err != nil {
		fmt.Println("Error when killing process")
		return err
	}
	compute_result(overall_result, sdc_main)

	overall_result, err = kill_process_by_name(stryker_api)
	if err != nil {
		fmt.Println("Error when killing process")
		return err
	}
	compute_result(overall_result, stryker_api)

	return nil
}

func kill_process_by_name(process_name string) ([]bool, error) {
	if process_name == "" {
		return nil, fmt.Errorf("invalid process name")
	}
	pids, err := get_all_processes_pids(process_name)
	if err != nil {
		return nil, err
	}

	var overall_result []bool
	for _, pid := range pids {
		result, err := kill_win_process(pid)
		if err != nil {
			overall_result = append(overall_result, false)
		}
		overall_result = append(overall_result, result)
	}

	return overall_result, nil
}

func get_all_processes_pids(target_process_name string) ([]uint32, error) {

	//var TH32CS_SNAPPROCESS uint32 = 0x02
	handle_32, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(handle_32)

	var process_entry windows.ProcessEntry32
	size_ptr := unsafe.Sizeof(process_entry)
	process_entry.Size = uint32(size_ptr)
	err = windows.Process32First(handle_32, &process_entry)
	if err != nil {
		return nil, err
	}

	var pids []uint32
	for {
		process_name := windows.UTF16ToString(process_entry.ExeFile[:])

		if strings.EqualFold(process_name, target_process_name) {
			pids = append(pids, process_entry.ProcessID)
		}

		err = windows.Process32Next(handle_32, &process_entry)
		if errors.Is(err, syscall.ERROR_NO_MORE_FILES) {
			break
		}
		if err != nil {
			return pids, fmt.Errorf("process32 next call failed")
		}
	}

	return pids, nil
}

func kill_win_process(pid uint32) (bool, error) {
	//const STILL_ACTIVE uint32 = 259

	handle32, win_err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if win_err != nil {
		return false, win_err
	}
	defer windows.CloseHandle(handle32)

	var exitCode uint32 = 1
	// if err := windows.GetExitCodeProcess(handle32, &exitCode); err != nil {
	// 	return false, fmt.Errorf("GetExitCodeProcess error %w", err)
	// } else if exitCode != STILL_ACTIVE {
	// 	return true, nil
	// }
	//windows.ExitProcess(exitCode)
	if err := windows.TerminateProcess(handle32, exitCode); err != nil {
		return false, fmt.Errorf("TerminateProcess error %w", err)
	}

	return true, nil
}

// func kill_process_pid(pid int) error {

// 	if pid <= 0 {
// 		return fmt.Errorf("invalid pid")
// 	}

// 	process, err := os.FindProcess(pid)
// 	if err != nil {
// 		return err
// 	}

// 	err = process.Kill()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
