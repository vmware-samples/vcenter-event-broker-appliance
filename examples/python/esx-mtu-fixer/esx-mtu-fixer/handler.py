import os
import ssl
import sys
import json
import atexit
import gzip, zlib
from pyVim import connect
from pyVmomi import vim


def get_vm_hosts(content, regex_esxi=None):
    host_view = content.viewManager.CreateContainerView(content.rootFolder,
                                                        [vim.HostSystem],
                                                        True)
    obj = [host for host in host_view.view]
    match_obj = []
    if regex_esxi:
        for esxi in obj:
            if re.findall(r'%s.*' % regex_esxi, esxi.name):
                match_obj.append(esxi)
        match_obj_name = [match_esxi.name for match_esxi in match_obj]
        host_view.Destroy()
        return match_obj
    else:
        host_view.Destroy()
        return obj


def handle(req):
    sslContext = ssl.SSLContext(ssl.PROTOCOL_SSLv23)
    sslContext.verify_mode = ssl.CERT_NONE

    service_instance = None
    
    vcenter_host = None
    vcenter_user = None
    vcenter_pass = None

    with open("/var/openfaas/secrets/vc-user","r") as vc_user:
        vcenter_user = vc_user.read()

    with open("/var/openfaas/secrets/vc-password","r") as vc_pass:
        vcenter_pass = vc_pass.read()

    with open("/var/openfaas/secrets/vc-host","r") as vc_host:
        vcenter_host = vc_host.read()

    try:
        service_instance = connect.SmartConnect(host=vcenter_host,
                                        user=vcenter_user,
                                        pwd=vcenter_pass,
                                        port=443,
                                        sslContext=sslContext)
        atexit.register(connect.Disconnect, service_instance)
    except IOError as e:
        sys.stderr.write(str(e))

    if not service_instance:
        sys.stderr.write(str("Unable to connect to host with supplied info."))
        
    esx_hosts = get_vm_hosts(service_instance.content)
    
    changes = "Changed hosts:\n"

    new_vnic = vim.host.VirtualNic.Specification()
    new_vnic.mtu = 1500

    for host in esx_hosts:
        for vnic in host.configManager.networkSystem.networkInfo.vnic:
            if vnic.spec.mtu < 1500:
                changes = changes + "host IP: " + host.name + "\nold mtu: " + str(vnic.spec.mtu) + "\nnew mtu: 1500\n"
                host.configManager.networkSystem.UpdateVirtualNic(vnic.device,new_vnic)   

    return changes
