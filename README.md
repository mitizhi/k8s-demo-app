0. This is a *pseudo web app*. Its sole **raison d'être** is to allow demonstrating
   some aspects of Kubernetes resources. It can be run as a pod, deployed as an
   independent service/ingress, or used as a reverse proxy target. To configure the
   behavior it can use ConfigMap values for the app configuration, employ
   secrets, and utilize constant and stateful volumes. It provides two suicidal
   scenarios for lifecycle demonstrations: abrupt middle-flight non-zero exit
   and delayed exit with zero exit value. It supports a pseudo-authenticated
   request (secret password) for retrieving a page (secret information), and
   implementes a simple integer counter for resulting in state change (counter).

1. The app uses **environmental variables** for basic **configuration**. The following
   ones affect the app behavior:

   | env variable | description                                                    |
   | ------------ | -------------------------------------------------------------- |
   | PORT         | Port to listen on (8080)                                       |
   | PREFIX       | URL path prefix (outside of that results to 404)               |
   | LOGLEVEL     | Log level (one of "debug", "info", "warning", "error", "fatal")|
   | SECRETPASSWD | Password required to acceass a page with "secret information"  |
   | SENSITIVEINFO| Something secret, only accessible if one knows the password    |
   | TEMPLATEDIR  | Template directory for populated simulated content             |
   | STATEDIR     | Stateful data directory (only "counter" there)                 |


2. Template directory is by default served from the container image under
   `data/` directory). Two kinds of volumes can be attached to it, one for the
   content and and volumes attached. The app by default comes with default
   payload as part of the container image, but can be configured to have that
   provided through a volume. This payload is not changed by the, but there is
   also a stateful data in the form of a counter that is by default stored
   inside the container image, but it is possible replace it, too, with
   a volume.

3. **Request URLs summarized:**


   | URL scheme                                                 | Description                                                                                                                           |
   |------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
   | http://${host}:${PORT}/${PREFIX}/                          | Service test URL for checking accessibility.                                                                                          |
   | http://${host}:${PORT}/${PREFIX}/version                   | Prints out the hard-coded version of the app.                                                                                         |
   | http://${host}:${PORT}/${PREFIX}/quit                      | Quits after waiting 1 seconds with status 0.                                                                                          |
   | http://${host}:${PORT}/${PREFIX}/crash                     | Abruptly crashes the service (status 1).                                                                                              |
   | http://${host}:${PORT}/${PREFIX}/count                     | Increase the counter value from a file on every request.                                                                              |
   | http://${host}:${PORT}/${PREFIX}/sensitive/${SECRETPASSWD} | Retrieve sensitive information (stored in ${SENSITIVEINFO}). If the  password is not correctly included then return 401 Unauthorized. |
   | http://${host}:${PORT}/${PREFIX}/${filename}               | Serves the template in ${TEMPLATEDIR}/${filename}.                                                                                    |
   | http://${host}:${PORT}/${PREFIX}/list/<host>               | List the supported URLs, putting <host> in place. Try `/list/localhost`.                                                              |

4. **Limitations**

   1. Host (either an IP address or corresponding host name) is not configurable.
	  The app always listen on `0.0.0.0:${PORT}`.
   2. While the app takes into account thread synchronization in handling the
	  counter incrementation (low hanging fruit), no effort is seen to handle
	  multi-container scenarios using the same state file. (For multi-node
	  solution one could apparently use some etcd service or perhaps database.)
