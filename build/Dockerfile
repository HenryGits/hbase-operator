FROM centos:7
LABEL REPOSITORY=zhc/hbase TAG=2.4.6

#RUN echo $' \n\
#[base]  \n\
#name=CentOS7-Base \n\
#baseurl=http://mirrors.163.com/centos/7/os/x86_64/ \n\
#gpgcheck=0 \n\
#gpgkey=http://mirrors.163.com/centos/RPM-GPG-KEY-CentOS-7' > /etc/yum.repos.d/CentOS-Base.repo
#RUN cat /etc/yum.repos.d/CentOS-Base.repo
#RUN yum clean all && yum makecache fast
#
#RUN yum install -y --nogpgcheck net-tools

# jdk依赖包
ENV JAVA_HOME=/usr/local/jdk
ENV CLASSPATH=.:${JAVA_HOME}/jre/lib/rt.jar:${JAVA_HOME}/lib/dt.jar:${JAVA_HOME}/lib/tools.jar

# Hbase环境变量配置
ENV HBASE_HOME=/usr/local/hbase
ENV CLASSPATH=$CLASSPATH:$HBASE_HOME/lib
ENV PATH=$PATH:$HBASE_HOME/bin

COPY  jdk /usr/local/jdk
COPY  hbase /usr/local/hbase
COPY  Hbase /Hbase

RUN chmod +x /Hbase

EXPOSE 16000 16020 9090

ENTRYPOINT ["/Hbase"]
