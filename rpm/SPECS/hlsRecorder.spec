%define version unknown
%define bin_name hlsRecorder
%define debug_package %{nil}

Name:           %{bin_name}
Version:        %{version}
Release:        1%{?dist}
Summary:        %{bin_name}: Record HLS
License:        BSD
URL:            http://git.itv.restr.im/infra/%{bin_name}
Source:         %{bin_name}-%{version}.tar.gz

%define restream_dir /opt/restream/
%define restream_bin_dir %{restream_dir}/%{bin_name}/bin

%description
This package provides hlsRecorder with verimatrix.

%prep
%setup

%build
make

%install
%{__mkdir} -p %{buildroot}%{restream_bin_dir}
%{__install} -m 0755 -p bin/%{bin_name} %{buildroot}%{restream_bin_dir}

%files
%defattr(-,root,root,-)
%{restream_bin_dir}/%{bin_name}
