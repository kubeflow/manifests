
feast/base: clean
	cd feast/base && helm template -f ../../values.yaml kf-feast feast --namespace feast --version 0.9.2 --repo https://storage.googleapis.com/feast-charts > resources.yaml

.PHONY:clean-kustomize
clean:
	rm -rf feast/base/resources.yaml