package selfupdater

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tomasen/rollover"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type SelfUpdater struct {
	LocalExecutablePath     string
	LocalExecutableChecksum string
	Provider                UpdateProvider
}

type UpdateProvider interface {
	// Download latest executable file from remote source to the open file descriptor provide by SelfUpdater.
	DownloadTo(file *os.File) error

	// Fetch the checksum of the executable file from remote source.
	RemoteChecksum() (string, error)

	// Return the hash.Hash interface that should be used to calculate the checksum of executable files.
	Hash() hash.Hash
}

func NewSelfUpdater(provider UpdateProvider) *SelfUpdater {
	ex, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}

	s := &SelfUpdater{
		LocalExecutablePath: ex,
		Provider:            provider,
	}
	s.UpdateLocalExecutableChecksum()
	return s
}

func (s *SelfUpdater) UpdateLocalExecutableChecksum() {
	s.LocalExecutableChecksum = s.CalcFileChecksum(s.LocalExecutablePath)
}

func (s *SelfUpdater) CalcFileChecksum(file string) string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	h := s.Provider.Hash()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *SelfUpdater) Update() error {
	// fetch remote checksum
	remoteChecksum, err := s.Provider.RemoteChecksum()
	if err != nil {
		return errors.WithMessage(err, "unable to obtain remote checksum")
	}

	if remoteChecksum == s.LocalExecutableChecksum {
		return nil
	}

	err = s.realUpdate(remoteChecksum)
	if err != nil {
		return err
	}

	s.Restart()
	return nil
}

func (s *SelfUpdater) realUpdate(remoteChecksum string) error {
	// download to tmpfile
	tmpFile, err := ioutil.TempFile(os.TempDir(), "self-update-executable-")
	if err != nil {
		return errors.WithMessage(err, "cannot create tmpfile")
	}

	//  clean up the file afterwards
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	s.Provider.DownloadTo(tmpFile)
	if err != nil {
		return errors.WithMessage(err, "unable to download remote executable")
	}

	tmpFile.Seek(0, 0)
	h := s.Provider.Hash()
	if _, err := io.Copy(h, tmpFile); err != nil {
		return errors.WithMessage(err, "error calculating checksum for tmpfile")
	}
	tempChecksum := fmt.Sprintf("%x", h.Sum(nil))

	if tempChecksum != remoteChecksum {
		return errors.New("checksum error on downloaded file")
	}

	info, err := os.Stat(s.LocalExecutablePath)
	log.Println(info.Mode())

	dest, err := os.OpenFile(s.LocalExecutablePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return errors.WithMessage(err, "error while opening current executable file for overwrite")
	}
	tmpFile.Seek(0, 0)
	buf := make([]byte, 4096)
	for {
		n, err := tmpFile.Read(buf)
		if err != nil && err != io.EOF {
			return errors.WithMessage(err, "read error while overwriting current executable file")
		}
		if n == 0 {
			break
		}
		if _, err := dest.Write(buf[:n]); err != nil {
			return errors.WithMessage(err, "write error while overwriting current executable file")
		}
	}
	if err := dest.Close(); err != nil {
		return errors.WithMessage(err, "error while closing current executable file")
	}

	return nil
}

func (s *SelfUpdater) Restart() {
	// Restart current running program
	p, err := rollover.Restart()
	if err != nil {
		log.Println("error rollover child", err)
	}
	if p == nil {
		log.Println("error starting child")
	} else {
		log.Println("child started running, pid:", p.Pid)
	}
}
