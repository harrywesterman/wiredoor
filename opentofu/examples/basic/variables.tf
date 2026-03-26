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

