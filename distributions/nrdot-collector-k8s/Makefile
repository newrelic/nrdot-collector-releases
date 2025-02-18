THIS_MAKEFILE_DIR := $(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sync-configs
sync-configs:
	@${THIS_MAKEFILE_DIR}/sync-configs.sh

.PHONY: check-k8s-configs
check-k8s-configs: sync-configs
	@git diff --name-only | grep -E 'distributions\/nrdot-collector-k8s\/config-\w*.yaml' \
    		&& { \
    			echo "Collector configs in helm chart have changed. Adjust sync script to take changes into account.";\
    			echo "Diff of $(NOTICE_OUTPUT):";\
    			git --no-pager diff HEAD -- 'distributions/nrdot-collector-k8s/config-*.yaml';\
    			exit 1;\
    		} \
    		|| exit 0
