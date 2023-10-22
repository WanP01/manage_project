package fs

import "os"

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) { // 等同errors.Is(err,fs.ErrExit) 判断该路径是否已有文件存在
			return true
		}
		return false
	}
	return true
}
