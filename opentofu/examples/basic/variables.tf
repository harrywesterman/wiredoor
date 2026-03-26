variable "endpoint" {
  type = string
}

variable "username" {
  type = string
}

variable "password" {
  type      = string
  sensitive = true
}

variable "insecure" {
  type    = bool
  default = false
}

variable "node_name" {
  type = string
}

variable "node_addr" {
  type = string
}

variable "domain_name" {
  type    = string
  default = ""
}

variable "domain_ssl" {
  type    = string
  default = "self-signed"
}

variable "http_name" {
  type = string
}

variable "http_path" {
  type    = string
  default = "/"
}

variable "http_port" {
  type = number
}

variable "tcp_backend_port" {
  type    = number
  default = 0
}

variable "tcp_enabled" {
  type    = bool
  default = false
}

variable "tcp_name" {
  type    = string
  default = ""
}

variable "tcp_port" {
  type    = number
  default = 0
}
