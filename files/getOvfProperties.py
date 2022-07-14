#!/usr/bin/env python3

from subprocess import Popen, PIPE
import sys
from xml.dom.minidom import parseString
import xml.parsers.expat
import re
import json
from os import path
from os.path import exists

if exists('/.dockerenv'):
    ovfenv_cmd = "/usr/bin/cat /root/setup/test/ovf-enf-test-1.xml"
else:
    ovfenv_cmd = "/usr/bin/vmtoolsd --cmd 'info-get guestinfo.ovfEnv'"
veba_config_file = "/root/config/veba-config.json"
veba_env_file = "/root/config/shell_env"


def debug(s):
    sys.stderr.write(s + " \n")  # Syserr only get logged on the console logs
    sys.stderr.flush()


def get_ovf_properties():
    """
        Return a dict of OVF properties in the ovfenv
    """
    properties = {}
    xml_parts = Popen(ovfenv_cmd, shell=True,
                      stdout=PIPE).stdout.read()
    try:
        raw_data = parseString(xml_parts)
    except xml.parsers.expat.ExpatError as err:
        debug(xml.parsers.expat.ErrorString(err.code))
        sys.exit(1)
    for property in raw_data.getElementsByTagName('Property'):
        key, value = [property.attributes['oe:key'].value,
                      property.attributes['oe:value'].value]
        properties[key] = value
    return properties


def remove_prefix(text, prefix):
    if text.startswith(prefix):
        return text[len(prefix):]
    return text


def main():
    ovf = get_ovf_properties()

    if path.isfile(veba_config_file):
        with open(veba_config_file) as fp:
            veba_config = json.load(fp)
    else:
        veba_config = {}

    with open(veba_env_file, 'w') as fp:
        for prop, res in ovf.items():
            property_name = remove_prefix(prop, 'guestinfo.')

            # Strip enclosing quotes if not a password
            if not re.search('password', property_name, flags=re.IGNORECASE):
                res = re.sub(r'^(["\'])(.*)\1$', r'\2', res)

            # Add the property to veba-config.json
            veba_config[property_name.upper()] = res

            # Escape special shell characters
            # res = re.sub("(!|\$|#|&|\"|\'|\(|\)|\||<|>|`|\\\|;)", r"\\\1", res)
            # res = re.sub(r'([$`"\\!])', r'\\\1', res)
            # res = re.sub("'", "'\"'\"'", res)
            res = re.sub('"', r'\"', res)

            # Write out the result to the environment variables
            fp.write(f'{property_name.upper()}="{res}"\n')

    with open(veba_config_file, 'w') as fp:
        json.dump(veba_config, fp, indent=4)

    sys.exit(0)


if __name__ == "__main__":
    main()
