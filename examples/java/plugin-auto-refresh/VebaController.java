package com.vmware.sample.remote.controllers;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import com.vmware.sample.remote.services.MessagingService;
import com.vmware.sample.remote.model.Message;
import com.vmware.sample.remote.model.MessageType;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping(value = "/veba")
public class VebaController {

    private static final Log logger = LogFactory.getLog(
            VebaController.class);

    private final MessagingService messagingService;

    public VebaController(final MessagingService messagingService) {
            this.messagingService = messagingService;
    }

    @RequestMapping(value = "/host/updated", method = RequestMethod.PUT)
    public void updateHostList(){
            messagingService.broadcastMessage(new Message(MessageType.HOSTS_LIST_UPDATED));
            logger.info("Message 'list of hosts updated' sent!");
    }
    
}