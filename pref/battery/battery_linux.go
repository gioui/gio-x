package battery

import (
	"fmt"
	"os"
	"path/filepath"
)

func batteryLevel() (uint8, error) {
	bat, err := newBattery()
	if err != nil {
		return 0, err
	}

	return bat.level()
}

func isSavingBattery() (bool, error) {
	const pattern = "/sys/devices/system/cpu/cpu*/cpufreq/scaling_governor"
	cpus, err := filepath.Glob(pattern)
	if err != nil || len(cpus) == 0 {
		return false, ErrNotAvailableAPI
	}

	raw, err := os.ReadFile(cpus[0])
	if err != nil {
		return false, ErrNotAvailableAPI
	}

	return string(raw) == "powersave\n", nil
}

func isCharging() (bool, error) {
	bat, err := newBattery()
	if err != nil {
		return false, err
	}
	return bat.isCharging()
}

type battery struct {
	root string
}

func newBattery() (battery, error) {
	const pattern = "/sys/class/power_supply/BAT*"
	bs, err := filepath.Glob(pattern)
	if err != nil || len(bs) == 0 {
		return battery{}, ErrNoSystemBattery
	}

	return battery{bs[0]}, nil
}

func (bat battery) level() (lvl uint8, err error) {
	if bat.root == "" {
		return 0, ErrNoSystemBattery
	}

	f, err := os.Open(filepath.Join(bat.root, "capacity"))
	if err != nil {
		return 0, ErrNotAvailableAPI
	}
	defer f.Close()

	_, err = fmt.Fscanf(f, "%d", &lvl)
	if err != nil {
		return 0, ErrNotAvailableAPI
	}

	return lvl, nil
}

func (bat battery) isCharging() (bool, error) {
	if bat.root == "" {
		return false, ErrNoSystemBattery
	}

	raw, err := os.ReadFile(filepath.Join(bat.root, "status"))
	if err != nil {
		return false, ErrNotAvailableAPI
	}

	return string(raw) == "charging\n", nil
}
