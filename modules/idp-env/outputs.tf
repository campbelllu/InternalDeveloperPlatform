output "instance_public_ip" {
  description = "The public IP address of the deployed developer sandbox"
  value = aws_instance.dev_node.public_ip
}

output "instance_id" {
  description = "The core AWS Instance ID used for SSM secure tunneling connection sessions"
  value = aws_instance.dev_node.id
}