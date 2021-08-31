ifndef $(version)
  version=$(shell git describe --tags `git rev-list --tags --max-count=1`)
endif

ARCH=$(shell [ `go env GOARCH`X == "amd64X" ] && echo x86_64 || echo i386)

all: reset-authentication

.PHONY: reset-authentication
reset-authentication:
	go build -o reset-authentication -ldflags "-X main.version=$(version)" -v ./cmd/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o reset-authentication.exe -ldflags "-X main.version=$(version)" -v ./cmd/main.go

.PHONY: rpm
rpm: reset-authentication
	rm -rf output || true
	mkdir output
	mkdir -p output/lib/systemd/system/ output/usr/local/bin/
	cp reset-authentication output/usr/local/bin/.
	cp conf/reset-authentication.service output/lib/systemd/system/.
	echo -e "systemctl enable reset-authentication.service\nsystemctl restart reset-authentication.service" > tmp_rpm.sh
	fpm -s dir -f -t rpm -a $(ARCH) --iteration 0 -n reset-authentication -v $(version) -C output --post-install tmp_rpm.sh
	rm -f tmp_rpm.sh

.PHONY: deb
deb: reset-authentication
	rm -rf output || true
	mkdir output
	mkdir -p output/lib/systemd/system/ output/usr/local/bin/
	cp reset-authentication output/usr/local/bin/.
	cp conf/reset-authentication.service output/lib/systemd/system/.
	echo -e "systemctl enable reset-authentication.service\nsystemctl restart reset-authentication.service" > tmp_rpm.sh
	fpm -s dir -f -t deb -a $(ARCH) --iteration 0 -n reset-authentication -v $(version) -C output --deb-no-default-config-files --post-install tmp_rpm.sh
	rm -f tmp_rpm.sh

.PHONY: rpm-for-redhat6
rpm-for-redhat6: reset-authentication
	rm -rf output || true
	mkdir output
	mkdir -p output/etc/init.d/ output/usr/local/bin/
	cp reset-authentication output/usr/local/bin/.
	cp conf/reset-authentication_for_redhat6.service output/etc/init.d/reset-authentication
	chmod 755 output/etc/init.d/reset-authentication
	echo -e "chkconfig reset-authentication on\nservice reset-authentication start" > tmp_rpm.sh
	fpm -s dir -f -t rpm -a $(ARCH) --iteration el6 -n reset-authentication -v $(version) -C output --post-install tmp_rpm.sh
	rm -f tmp_rpm.sh

