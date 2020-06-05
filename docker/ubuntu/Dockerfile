FROM ubuntu:latest

RUN apt update
RUN apt install -y nano vim sudo 
RUN apt install -y openssh-server
RUN mkdir /var/run/sshd

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

COPY cont_init.sh /
EXPOSE 22
CMD ["/usr/sbin/sshd"]
