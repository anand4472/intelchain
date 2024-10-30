{{VER=2.0.0}}
{{REL=0}}
# SPEC file overview:
# https://docs.fedoraproject.org/en-US/quick-docs/creating-rpm-packages/#con_rpm-spec-file-overview
# Fedora packaging guidelines:
# https://docs.fedoraproject.org/en-US/packaging-guidelines/

Name:		intelchain
Version:	{{ VER }}
Release:	{{ REL }}
Summary:	intelchain blockchain validator node program

License:	MIT
URL:		https://intelchain.org
Source0:	%{name}-%{version}.tar
BuildArch: x86_64
Packager: winay jadon <winay@intelchain.org>
Requires(pre): shadow-utils
Requires: systemd-rpm-macros jq

%description
Intelchain is a sharded, fast finality, low fee, PoS public blockchain.
This package contains the validator node program for intelchain blockchain.

%global debug_package %{nil}

%prep
%setup -q

%build
exit 0

%check
./intelchain --version
exit 0

%pre
getent group intelchain >/dev/null || groupadd -r intelchain
getent passwd intelchain >/dev/null || \
   useradd -r -g intelchain -d /home/intelchain -m -s /sbin/nologin \
   -c "Intelchain validator node account" intelchain
mkdir -p /home/intelchain/.itc/blskeys
mkdir -p /home/intelchain/.config/rclone
chown -R intelchain.intelchain /home/intelchain
exit 0


%install
install -m 0755 -d ${RPM_BUILD_ROOT}/usr/sbin ${RPM_BUILD_ROOT}/etc/systemd/system ${RPM_BUILD_ROOT}/etc/sysctl.d ${RPM_BUILD_ROOT}/etc/intelchain
install -m 0755 -d ${RPM_BUILD_ROOT}/home/intelchain/.config/rclone
install -m 0755 intelchain ${RPM_BUILD_ROOT}/usr/sbin/
install -m 0755 intelchain-setup.sh ${RPM_BUILD_ROOT}/usr/sbin/
install -m 0755 intelchain-rclone.sh ${RPM_BUILD_ROOT}/usr/sbin/
install -m 0644 intelchain.service ${RPM_BUILD_ROOT}/etc/systemd/system/
install -m 0644 intelchain-sysctl.conf ${RPM_BUILD_ROOT}/etc/sysctl.d/99-intelchain.conf
install -m 0644 rclone.conf ${RPM_BUILD_ROOT}/etc/intelchain/
install -m 0644 intelchain.conf ${RPM_BUILD_ROOT}/etc/intelchain/
exit 0

%post
%systemd_user_post %{name}.service
%sysctl_apply %{name}-sysctl.conf
exit 0

%preun
%systemd_user_preun %{name}.service
exit 0

%postun
%systemd_postun_with_restart %{name}.service
exit 0

%files
/usr/sbin/intelchain
/usr/sbin/intelchain-setup.sh
/usr/sbin/intelchain-rclone.sh
/etc/sysctl.d/99-intelchain.conf
/etc/systemd/system/intelchain.service
/etc/intelchain/intelchain.conf
/etc/intelchain/rclone.conf
/home/intelchain/.config/rclone

%config(noreplace) /etc/intelchain/intelchain.conf
%config /etc/intelchain/rclone.conf
%config /etc/sysctl.d/99-intelchain.conf 
%config /etc/systemd/system/intelchain.service

%doc
%license



%changelog
* Wed Aug 26 2020 Winay Jadon <winay at intelchain dot org> 2.3.5
   - get version from git tag
   - add %config macro to keep edited config files

* Tue Aug 18 2020 Winay Jadon <winay at intelchain dot org> 2.3.4
   - init version of the intelchain node program

