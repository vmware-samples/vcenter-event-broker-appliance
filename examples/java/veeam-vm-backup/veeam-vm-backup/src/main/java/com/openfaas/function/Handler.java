// The source of the code in  the getUnsafeOkHttpClient() method
// can be found here: https://gist.github.com/drakeet/93be3c6ff9ce2d4ca8fb
package com.openfaas.function;

import com.moandjiezana.toml.Toml;
import com.openfaas.model.IRequest;
import com.openfaas.model.IResponse;
import com.openfaas.model.Response;
import okhttp3.*;
import org.json.JSONObject;
import org.w3c.dom.Document;
import org.w3c.dom.Node;
import org.w3c.dom.NodeList;
import org.xml.sax.InputSource;
import org.xml.sax.SAXException;

import javax.net.ssl.*;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.parsers.ParserConfigurationException;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.StringReader;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.util.Base64;
import java.util.StringTokenizer;


public class Handler extends com.openfaas.model.AbstractHandler implements com.openfaas.model.IHandler {
    // VEEAM's RESTful API endpoints
    private static final String API_LOGON_SESSION_PATH = "/api/sessionMngr/?v=latest";
    private static final String API_HIERARCHY_ROOTS_PATH = "/api/hierarchyRoots";
    private static final String API_BACKUP_PATH  = "/api/backupServers/";
    private static final String API_BACKUP_JOB = "?action=quickbackup";
    private static final String VEEAM_CONFIG = "/var/openfaas/secrets/veeamconfig";

    public IResponse Handle(IRequest req) {
        // Get the data from toml file
        String ip = null;
        String port = null;
        String id = null;
        String vcIp = null;
        String encodedCredentials = null;

        try {
            System.out.println("Reading toml configuration data started");

            InputStream inputStream = new FileInputStream(VEEAM_CONFIG);
            Toml toml = new Toml().read(inputStream);

            ip = toml.getString("enterprise_manager.ip");
            port = toml.getString("enterprise_manager.port");
            String username = toml.getString("enterprise_manager.user");
            String password = toml.getString("enterprise_manager.password");
            id = toml.getString("backup_server.id");
            vcIp = toml.getString("vcenter.ip");
            System.out.println("Reading toml configuration data finished");

            // Encode credentials
            String credentials = username + ":" + password;
            encodedCredentials = Base64.getEncoder().encodeToString(credentials.getBytes());
            System.out.println("Credentialds encoded!");

        }   catch (Exception exception){
            exception.printStackTrace();
        }

        OkHttpClient client = getUnsafeOkHttpClient();
        
        // Create a logon session to the Veeam server using the credentials from the toml file, which are now encoded
        okhttp3.Response logonRes = logonSession(ip, port, encodedCredentials, client);

        // Obtain the session ID from the logon response. 
        // The session ID is an identification number of the logon session.
        String sessionID = logonRes.header("X-RestSvcSessionId");
        System.out.println("Session ID saved!");
        
        // Veeam server uses its own ids for the hosts in its managed servers. 
        // Hierarchy roots represent a collection of all virtualization hosts added to the Veeam backup servers 
        // connected to Veeam Backup Enterprise Manager
        okhttp3.Response hierarchyRoots = getHierarchyRoots(ip, port, sessionID, client);
        String xmlResponseBody = null;
        try {
            xmlResponseBody = hierarchyRoots.body().string();
        } catch (IOException e) {
            e.printStackTrace();
        }
        String vcId = getVcId(xmlResponseBody, vcIp);

        // Construct the vm reference
        // The VM reference used by Veeam is different from the one defined in the vCenter
        String vmRef = getVmRef(req, vcId);
        
        // Make a backup request to the Veeam server
        IResponse backupResponse = backupRequest(vmRef, ip, port, id, sessionID, client);

        //Return the response from the backup request
        return backupResponse;
    }


    private String getVmRef(IRequest req, String vcId){
        //Parse the IRequest to JSON and get the VM's info
        String jsonString = req.getBody();
        JSONObject obj = new JSONObject(jsonString);
        String vmValue = obj.getJSONObject("data").getJSONObject("Vm").getJSONObject("Vm").getString("Value");

        String vmRef = "urn:VMware:vm:" + vcId + "." + vmValue;
        System.out.println("VM reference received!");

        return vmRef;
    }

    private String getVcId(String xmlResponseBody, String vcIp){
        DocumentBuilder dBuilder = null;
        try {
            dBuilder = DocumentBuilderFactory.newInstance().newDocumentBuilder();
        } catch (ParserConfigurationException e) {
            e.printStackTrace();
        }

        String vcId = null;
        InputSource xml;
        try {
            xml = new InputSource(new StringReader(xmlResponseBody));
            Document doc = dBuilder.parse(xml);
            NodeList references = doc.getElementsByTagName("Ref");
            int index = -1;
            boolean found = false;
            do {
                index++;
                Node currentRef = references.item(index);
                String currentVcIp = currentRef.getAttributes().getNamedItem("Name").getNodeValue();
                if(vcIp.equals(currentVcIp)){
                    found = true;
                }
            }while(index < references.getLength() && !found);

            if(found) {
                String vcUrn = references.item(index).getAttributes().getNamedItem("UID").getNodeValue();
                StringTokenizer st = new StringTokenizer(vcUrn,":");
                while (st.hasMoreTokens()) {
                    vcId = st.nextToken();
                }
            }
        } catch (SAXException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
        return vcId;
    }


    private okhttp3.Response logonSession(String ip, String port,
                                          String encodedCredentials,  OkHttpClient client){
        okhttp3.Response logonRes = null;
        try {
            //Build logon request to VEEAM
            Request logonRequest = new Request.Builder()
                    .url("https://"+ ip + ":" + port + API_LOGON_SESSION_PATH)
                    .addHeader("Authorization", "Basic " + encodedCredentials)
                    .post(new FormBody.Builder().build())
                    .build();
            System.out.println("Logon request created!");

            //Send the logon request
            logonRes = client.newCall(logonRequest).execute();
            System.out.println("Logon request sent!");

        } catch (Exception e){
            e.printStackTrace();
            System.out.println(e.toString());
        }
        return logonRes;
    }

    private okhttp3.Response getHierarchyRoots(String ip, String port, String sessionID, OkHttpClient client){

        okhttp3.Response hierarchyRootsRes = null;
        try {
            //Create the hierarchy roots request
            Request hierarchyRootsRequest = new Request.Builder()
                    .url("https://" + ip + ":" + port + API_HIERARCHY_ROOTS_PATH)
                    .addHeader("X-RestSvcSessionId", sessionID)
                    .get()
                    .build();
            System.out.println("Hierarchy roots request created!");
            
            //Send the backup request
            hierarchyRootsRes = client.newCall(hierarchyRootsRequest).execute();
            System.out.println("Hierarchy roots request sent!");

        } catch (Exception e){
            e.printStackTrace();
            System.out.println(e.toString());
        }
        return hierarchyRootsRes;
    }

    private IResponse backupRequest(String vmRef, String ip, String port,
                                    String id, String sessionID, OkHttpClient client){

        Response backupRes = new Response();
        try {
            //Create the body for the backup request
            MediaType mediaType = MediaType.parse("text/xml");
            String body = "<?xml version=\"1.0\" " +
                    "encoding=\"utf-8\"?>\n " +
                    "<QuickBackupStartupSpec xmlns=\"http://www.veeam.com/ent/v1.0\" " +
                    "xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" " +
                    "xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\">\n" +
                    " <VmRef>" + vmRef + "</VmRef>\n" +
                    "</QuickBackupStartupSpec>";
            RequestBody formBody = RequestBody.create(mediaType, body);
            System.out.println("Backup request body is done!");

            //Create the backup request
            Request backupRequest = new Request.Builder()
                    .url("https://" + ip + ":" + port + API_BACKUP_PATH + id + API_BACKUP_JOB)
                    .addHeader("X-RestSvcSessionId", sessionID)
                    .addHeader("Content-Type", "application/xml")
                    .post(formBody)
                    .build();
            System.out.println("Backup request created!");

            //Send the backup request
            okhttp3.Response response = client.newCall(backupRequest).execute();
            System.out.println("Backup request sent!");

            backupRes.setBody(response.body().string());
           
        } catch (Exception e){
            e.printStackTrace();
            System.out.println(e.toString());
        }
        return backupRes;
    }


    //Ignore the problems with the self-signed certificate
    private static OkHttpClient getUnsafeOkHttpClient() {
        try {
            // Create a trust manager that does not validate certificate chains
            final TrustManager[] trustAllCerts = new TrustManager[]{
                    new X509TrustManager() {
                        @Override
                        public void checkClientTrusted(java.security.cert.X509Certificate[] chain,
                                                       String authType) throws CertificateException {
                        }
                        @Override
                        public void checkServerTrusted(java.security.cert.X509Certificate[] chain,
                                                       String authType) throws CertificateException {
                        }
                        @Override
                        public java.security.cert.X509Certificate[] getAcceptedIssuers() {
                            return new X509Certificate[0];
                        }
                    }
            };
            // Install the all-trusting trust manager
            final SSLContext sslContext = SSLContext.getInstance("SSL");
            sslContext.init(null, trustAllCerts, new java.security.SecureRandom());
            // Create an ssl socket factory with our all-trusting manager
            final SSLSocketFactory sslSocketFactory = sslContext.getSocketFactory();
            return new OkHttpClient.Builder()
                    .sslSocketFactory(sslSocketFactory, (X509TrustManager) trustAllCerts[0])
                    .hostnameVerifier(new HostnameVerifier() {
                        @Override
                        public boolean verify(String hostname, SSLSession session) {
                            return true;
                        }
                    }).build();
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}