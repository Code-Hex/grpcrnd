package call

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
)

func (c *command) output(result string) error {
	if c.uselog {
		os.Mkdir("logs", 0755) // through the error
		p, err := filepath.Abs("logs")
		if err != nil {
			return errors.Wrap(err, "failed to get full path")
		}
		logPath := filepath.Join(p, "response_log")
		logf, err := rotatelogs.New(
			logPath+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(logPath),
			rotatelogs.WithMaxAge(24*time.Hour),
			rotatelogs.WithRotationTime(time.Hour),
		)
		if err != nil {
			return errors.Wrap(err, "failed to create rotatelogs")
		}
		io.Copy(logf, strings.NewReader(result+"\n"))
	} else {
		fmt.Println(result)
	}
	return nil
}
