package readers

import (
	"bufio"
	//"fmt"
	//"net/url"
	"os"
	//"strconv"
	//"strings"
	"net/mail"

	//"github.com/helviojunior/sprayshark/internal/islazy"
)

// FileReader is a reader that expects a file with targets that
// is newline delimited.
type FileReader struct {
	Options *FileReaderOptions
}

// FileReaderOptions are options for the file reader
type FileReaderOptions struct {
	UserFile    string
	PassFile	string
}

// NewFileReader prepares a new file reader
func NewFileReader(opts *FileReaderOptions) *FileReader {
	return &FileReader{
		Options: opts,
	}
}

// Read from a file that contains targets.
// FilePath can be "-" indicating that we should read from stdin.
func (fr *FileReader) ReadPasswords(passwords *[]string) error {

	var file *os.File
	var err error

	file, err = os.Open(fr.Options.PassFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		candidate := scanner.Text()
		if candidate == "" {
			continue
		}

		*passwords = append(*passwords, candidate)
	}

	return scanner.Err()
}

// Read from a file that contains targets.
// FilePath can be "-" indicating that we should read from stdin.
func (fr *FileReader) ReadEmails(users *[]string)  error {

	var file *os.File
	var err error

	if fr.Options.UserFile == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(fr.Options.UserFile)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		candidate := scanner.Text()
		if candidate == "" {
			continue
		}

		m, err := mail.ParseAddress(candidate)
		if err == nil {
			*users = append(*users, m.Address)
		}
	}

	return scanner.Err()
}
