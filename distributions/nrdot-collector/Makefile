THIS_MAKEFILE_DIR := $(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

.PHONY: assert-config-invariants
assert-config-invariants:
	@${THIS_MAKEFILE_DIR}/assert-invariants-agent-control.sh
