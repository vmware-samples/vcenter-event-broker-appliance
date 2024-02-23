packer {
  required_version = ">= 1.8.5"
  required_plugins {
    vmware = {
      version = ">= v1.0.7"
      source  = "github.com/hashicorp/vmware"
    }
  }
}

source "vmware-iso" "veba" {
  boot_command         = ["<esc><wait>", "vmlinuz initrd=initrd.img root=/dev/ram0 loglevel=3 ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/photon-kickstart.json photon.media=cdrom insecure_installation=1", "<enter>"]
  boot_wait            = "10s"
  disk_additional_size = ["25600"]
  disk_size            = "25600"
  disk_type_id         = "thin"
  disk_adapter_type    = "pvscsi"
  format               = "ovf"
  guest_os_type        = "vmware-photon-64"
  headless             = false
  http_directory       = "http"
  insecure_connection  = true
  iso_checksum         = "${var.iso_checksum}"
  iso_url              = "${var.iso_url}"
  remote_datastore     = "${var.builder_host_datastore}"
  remote_host          = "${var.builder_host}"
  remote_password      = "${var.builder_host_password}"
  remote_type          = "esx5"
  remote_username      = "${var.builder_host_username}"
  shutdown_command     = "/sbin/shutdown -h now"
  shutdown_timeout     = "1000s"
  skip_compaction      = true
  ssh_password         = "${var.guest_password}"
  ssh_port             = 22
  ssh_username         = "${var.guest_username}"
  version              = "17"
  vm_name              = "${var.vm_name}"
  cpus                 = "${var.numvcpus}"
  memory               = "${var.ramsize}"
  network_adapter_type = "vmxnet3"
  network_name         = "${var.builder_host_portgroup}"
  vmx_data = {
    annotation = "Version: ${var.VEBA_VERSION}"
  }
  vnc_over_websocket = true
}

build {
  sources = ["source.vmware-iso.veba"]

  provisioner "shell" {
    inline = ["mkdir -p /root/config && mkdir -p /root/download"]
  }

  provisioner "file" {
    destination = "/root/config/veba-bom.json"
    source      = "veba-bom.json"
  }

  provisioner "file" {
    destination = "/root/download/"
    source      = "files/downloads/"
  }

  provisioner "shell" {
    environment_vars  = ["VEBA_VERSION=${var.VEBA_VERSION}", "VEBA_COMMIT=${var.VEBA_COMMIT}"]
    expect_disconnect = true
    scripts           = ["scripts/photon-settings.sh", "scripts/photon-docker.sh"]
  }

  provisioner "shell" {
    environment_vars = ["VEBA_VERSION=${var.VEBA_VERSION}"]
    pause_before     = "20s"
    scripts          = ["scripts/photon-containers.sh", "scripts/photon-cleanup.sh"]
  }

  provisioner "file" {
    destination = "/etc/rc.d/rc.local"
    source      = "files/rc.local"
  }

  provisioner "file" {
    destination = "/root/setup/getOvfProperty.py"
    source      = "files/getOvfProperty.py"
  }

  provisioner "file" {
    destination = "/root/setup/setup.sh"
    source      = "files/setup.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-01-os.sh"
    source      = "files/setup-01-os.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-02-proxy.sh"
    source      = "files/setup-02-proxy.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-03-network.sh"
    source      = "files/setup-03-network.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-04-kubernetes.sh"
    source      = "files/setup-04-kubernetes.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-05-knative.sh"
    source      = "files/setup-05-knative.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-06-vsphere-sources.sh"
    source      = "files/setup-06-vsphere-sources.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-07-event-router-webhook.sh"
    source      = "files/setup-07-event-router-webhook.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-08-tinywww.sh"
    source      = "files/setup-08-tinywww.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-09-ingress.sh"
    source      = "files/setup-09-ingress.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-010-veba-ui.sh"
    source      = "files/setup-010-veba-ui.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-011-fluentbit.sh"
    source      = "files/setup-011-fluentbit.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-012-cadvisor.sh"
    source      = "files/setup-012-cadvisor.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-098-dcui-endpoints.sh"
    source      = "files/setup-098-dcui-endpoints.sh"
  }

  provisioner "file" {
    destination = "/root/setup/setup-099-banner.sh"
    source      = "files/setup-099-banner.sh"
  }

  provisioner "file" {
    destination = "/usr/bin/veba-dcui"
    source      = "files/veba-dcui"
  }

  provisioner "file" {
    destination = "/boot/grub/themes/photon/photon.png"
    source      = "logo/veba_icon_only.png"
  }

  provisioner "file" {
    destination = "/root/config/"
    source      = "files/configs/"
  }

  post-processor "shell-local" {
    environment_vars = ["VEBA_VERSION=${var.VEBA_VERSION}", "VEBA_APPLIANCE_NAME=${var.vm_name}", "FINAL_VEBA_APPLIANCE_NAME=${var.vm_name}_${var.VEBA_VERSION}", "VEBA_OVF_TEMPLATE=${var.veba_ovf_template}"]
    inline           = ["cd manual", "./add_ovf_properties.sh"]
  }
  post-processor "shell-local" {
    inline = ["pwsh -F unregister_vm.ps1 ${var.ovftool_deploy_vcenter} ${var.ovftool_deploy_vcenter_username} ${var.ovftool_deploy_vcenter_password} ${var.vm_name}"]
  }
}
