// SPDX-License-Identifier: Unlicense OR MIT

#include "textflag.h"

TEXT ·fileRead(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·fileWrite(SB), NOSPLIT, $0
  CallImport
  RET
