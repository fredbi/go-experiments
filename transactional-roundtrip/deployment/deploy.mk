CHART_PATH=$(ROOT)/deployment

helm-debug:
	helm template \
		--debug \
		-f ${CHART_PATH}/values.yaml \
		--set-file config=${ROOT}/config.yaml \
		--set image.tag=${TAG} \
		${APP} \
		${CHART_PATH}

helm-deploy:
	helm upgrade \
		--install \
		-f ${CHART_PATH}/values.yaml \
		--set-file config=${ROOT}/config.yaml \
		--set image.tag=${TAG} \
		${APP} \
		${CHART_PATH}

helm-undeploy:
	helm uninstall ${APP}
