FROM centos:6

RUN echo -e "[wandisco-Git]\nname=CentOS-6 - Wandisco Git\nbaseurl=http://opensource.wandisco.com/centos/6/git/\$basearch/\nenabled=1\ngpgcheck=0" > /etc/yum.repos.d/wandisco-git.repo && \
    yum install -y epel-release && \
    yum install -y wget make git gzip rpm-build nc && \
    yum groupinstall -y "Development tools" && \
    yum clean all

RUN wget https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz -O /tmp/go.linux-amd64.tar.gz && \
    tar xvf /tmp/go.linux-amd64.tar.gz -C /usr/local && \
    ln -s /usr/local/go/bin/go* /usr/local/bin/ && \
    rm -f /tmp/go.linux-amd64.tar.gz
