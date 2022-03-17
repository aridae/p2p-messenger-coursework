package wszefeer2peer

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	STATIC_PATH = "./frontend/build"
)

type StaticServer struct {
	staticPath string
}

func NewStaticServer() *StaticServer {
	return &StaticServer{
		staticPath: STATIC_PATH,
	}
}

func (server *StaticServer) ProcessStaticRequest(request *http.Request, response *http.Response) {
	path := path.Clean(request.URL.Path)
	log.Printf("Request: %v\n", path)
	filePath := server.staticPath + path

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		response.StatusCode = 404
		response.Body = ioutil.NopCloser(strings.NewReader("Not found!"))
		return
	}

	if info.IsDir() {
		_, err := os.Stat(filePath + "index.html")
		if err == nil {
			responseFile(response, filePath+"index.html")
			return
		}

		files, err := readDir(filePath)
		if err != nil {
			response.StatusCode = 500
			response.Body = ioutil.NopCloser(strings.NewReader("Internal server error: " + err.(*os.PathError).Err.Error()))
		}
		filesString := strings.Join(files[:], "\n")
		response.Body = ioutil.NopCloser(strings.NewReader("Index of " + path + ":\n\n" + filesString))
		return
	}
	responseFile(response, filePath)
}

// Отдает содержимое файла
func responseFile(response *http.Response, fileName string) {
	file, err := os.Open(fileName)

	if os.IsPermission(err) {
		response.StatusCode = 403
		response.Body = ioutil.NopCloser(strings.NewReader("Forbidden"))
		return
	} else if err != nil {
		response.StatusCode = 500
		response.Body = ioutil.NopCloser(strings.NewReader("Internal server error: " + err.(*os.PathError).Err.Error()))
		return
	}

	response.Body = file
}

// Отдает список содержимого каталога
func readDir(root string) ([]string, error) {
	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		if file.IsDir() {
			files = append(files, file.Name()+"/")
		} else {
			files = append(files, file.Name())
		}
	}
	return files, nil
}
