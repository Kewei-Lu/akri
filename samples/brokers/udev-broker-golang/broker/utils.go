package broker

import (
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
)

const DEVPATH_REGEX string = "UDEV_DEVNODE_[A-F0-9]{6,6}=(.*)$"

var (
	Logger  *logrus.Logger = logrus.New()
	DEVPATH string
)

func InitLogger() (*logrus.Logger, error) {
	Logger.SetFormatter(&logrus.TextFormatter{})
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.DebugLevel) // for debug, set DEBUG LEVEL
	return Logger, nil
}

func GetDevPath() error {
	envs := os.Environ()
	re, err := regexp.Compile(DEVPATH_REGEX)
	if err != nil {
		Logger.Errorf("error in REGEX MATCH, err: %s", err.Error())
		// TODO: should not only return nil
		return nil
	}
	for _, e := range envs {
		if re.MatchString(e) {
			subMatch := re.FindStringSubmatch(e)
			if len(subMatch) < 2 {
				Logger.Errorf("error in SubMatch ENV DEVPATH, subMatch: %v", subMatch)
			}
			Logger.Debugf("DEVPATH: %s", subMatch[1])
			DEVPATH = subMatch[1]
		}

	}
	return nil
}
