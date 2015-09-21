This is a proxy between data received from collectd write_http plugin
and opentsdb. Edit the template file and save it as
collectd2tsdb.json. You can then run the program. The only
command-line option is '-c' to define a non-standard config file.

Your collectd server should be configured to send its measurements to
the proxy by adding the following lines in collectd.conf:

LoadPlugin write_http
<Plugin "write_http">
        <URL "http://10.0.0.1:8000/opentsdb">
             Format JSON
        </URL>
</Plugin>

Released under the WTFPL v2.0
