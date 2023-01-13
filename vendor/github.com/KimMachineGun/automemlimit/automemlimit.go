package automemlimit

import (
	"github.com/KimMachineGun/automemlimit/memlimit"
)

func init() {
	memlimit.SetGoMemLimitWithEnv()
}
