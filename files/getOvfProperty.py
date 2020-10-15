#!/usr/bin/env python3

import subprocess
import sys
from xml.dom.minidom import parseString

ovfenv_cmd="/usr/bin/vmtoolsd --cmd 'info-get guestinfo.ovfEnv'"


def debug(s):
    sys.stderr.write(s + " \n")  # Syserr only get logged on the console logs
    sys.stderr.flush()


def get_ovf_properties():
    """
        Return a dict of OVF properties in the ovfenv
    """
    properties = {}
    xml_parts = subprocess.Popen(ovfenv_cmd, shell=True,
                                 stdout=subprocess.PIPE).stdout.read()
    try:
        raw_data = parseString(xml_parts)
    except xml.parsers.expat.ExpatError as err:
        debug(e)
        sys.exit(1)
    for property in raw_data.getElementsByTagName('Property'):
        key, value = [ property.attributes['oe:key'].value,
                       property.attributes['oe:value'].value ]
        properties[key] = value
    return properties


def main(argv):
    if len(argv) is not 1:
        debug('usage: getOvfProperty.py <property_name')
        sys.exit(1)
           
    ovf = get_ovf_properties()
    
    try:
        res = ovf[argv[0]]
    except KeyError as err:
        debug(f'ovfProperty not found: {err}')
        sys.exit(1)
    
    print(res, end="")
    sys.exit(0)


if __name__ == "__main__":
   main(sys.argv[1:])