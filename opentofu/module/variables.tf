variable "node_name" {
  type = string
}

variable "node_addr" {
  type = string
}

variable "allow_internet" {
  type    = bool
  default = false
}

variable "node_enabled" {
  type    = bool
  default = true
}

variable "pat_name" {
  type    = string
  default = "default"
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

variable "http_backend_host" {
  type    = string
  default = ""
}

variable "http_backend_proto" {
  type    = string
  default = "http"
}

variable "http_domain" {
  type    = string
  default = ""
}

variable "http_enabled" {
  type    = bool
  default = true
}

variable "http_require_auth" {
  type    = bool
  default = false
}

