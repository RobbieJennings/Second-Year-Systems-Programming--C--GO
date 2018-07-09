#include "option_handler.h"
#include <stdio.h>
#include <stdlib.h>
#include <fstream>
#include <iostream>
#include <string>
#include <string.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <netdb.h>

class client
{
private:
	int sock;
	struct sockaddr_in server;
public:
	client(std::string, int);
	void send_data(std::string);
	std::string receive_data(int);
};	

client::client(std::string host, int port)
{
	sock = socket(AF_INET, SOCK_STREAM, 0);
	struct hostent *he;
	struct in_addr **addr_list;

	he = gethostbyname(host.c_str());
	addr_list = (struct in_addr **) he->h_addr_list;
	server.sin_addr = *addr_list[0];
	server.sin_family = AF_INET;
	server.sin_port = htons(port);
	connect(sock, (struct sockaddr *)&server, sizeof(server));
}

void client::send_data(std::string data)
{
	send(sock, data.c_str(), strlen(data.c_str()), 0);
}

std::string client::receive_data(int size = 512)
{
	char buffer[size];
	bzero(buffer, size);
	recv(sock, buffer, sizeof(buffer), 0);
	return buffer;
}

int main(int argc, char* argv[])
{	
	std::string host;
	int port;
	bool web;
	std::ofstream file;

	OptionHandler::Handler h = OptionHandler::Handler(argc, argv);
	try
	{
		h.add_option('h', "host", OptionHandler::REQUIRED, false);
	}
	catch(const std::exception & e)
	{
		std::cerr << e.what() << std::endl;
	}
        try
        {
                h.add_option('p', "port", OptionHandler::REQUIRED, false);
        }
        catch(const std::exception & e)
        {
                std::cerr << e.what() << std::endl;
        }
        try
        {
                h.add_option('w', "web", OptionHandler::NONE, false);
        }
        catch(const std::exception & e)
        {
                std::cerr << e.what() << std::endl;
        }
        try
        {
                h.add_option('f', "file", OptionHandler::REQUIRED, false);
        }
        catch(const std::exception & e)
        {
                std::cerr << e.what() << std::endl;
        }
	try
        {
                h.add_option('H', "help", OptionHandler::NONE, false);
        }
        catch(const std::exception & e)
        {
                std::cerr << e.what() << std::endl;
        }
	try
        {
                h.add_option('?', "help", OptionHandler::NONE, false);
        }
        catch(const std::exception & e)
        {
                std::cerr << e.what() << std::endl;
        }
	
	if(h.get_option("help"))
	{
		std::cout << "usage: ./knock -h host -p port [-H] [-w] [-f file]\n";
	}
	else
	{
		if(h.get_option("host") && h.get_option("port"))
		{
			host = h.get_arguments("host").at(0);
			port = atoi(h.get_arguments("port").at(0).c_str());
			client c(host, port);
			if(h.get_option("web"))
			{
				c.send_data("GET / \r\n\r\n");        
			}
			else
			{
				c.send_data("GET \r\n\r\n");
			}
			std::string response = c.receive_data(1024);
			if(h.get_option("file"))
                        {
				file.open(h.get_arguments("file").at(0).c_str());
				file<< response;
               		}
		}
		else
		{
			std::cout << "Invalid input\nusage: ./knock -h host -p port [-H] [-w] [-f file]\n";
		}
	}
}
