#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

OUTPUT_PATH="../output-vmware-iso"

rm -f ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.mf

sed "s/{{VERSION}}/${VEBA_VERSION}/g" ${VEBA_OVF_TEMPLATE} > photon.xml

if [ "$(uname)" == "Darwin" ]; then
    sed -i .bak1 's/<VirtualHardwareSection>/<VirtualHardwareSection ovf:transport="com.vmware.guestInfo">/g' ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i .bak2 "/    <\/vmw:BootOrderSection>/ r photon.xml" ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i .bak3 '/^      <vmw:ExtraConfig ovf:required="false" vmw:key="nvram".*$/d' ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i .bak4 "/^    <File ovf:href=\"${VEBA_APPLIANCE_NAME}-file1.nvram\".*$/d" ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i .bak5 '/vmw:ExtraConfig.*/d' ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
else
    sed -i 's/<VirtualHardwareSection>/<VirtualHardwareSection ovf:transport="com.vmware.guestInfo">/g' ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i "/    <\/vmw:BootOrderSection>/ r photon.xml" ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i '/^      <vmw:ExtraConfig ovf:required="false" vmw:key="nvram".*$/d' ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i "/^    <File ovf:href=\"${VEBA_APPLIANCE_NAME}-file1.nvram\".*$/d" ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
    sed -i '/vmw:ExtraConfig.*/d' ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf
fi

ovftool ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}/${VEBA_APPLIANCE_NAME}.ovf ${OUTPUT_PATH}/${FINAL_VEBA_APPLIANCE_NAME}.ova
rm -rf ${OUTPUT_PATH}/${VEBA_APPLIANCE_NAME}
rm -f photon.xml
