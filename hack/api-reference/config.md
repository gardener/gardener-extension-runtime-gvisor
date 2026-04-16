<p>Packages:</p>
<ul>
<li>
<a href="#gvisor.runtime.extensions.config.gardener.cloud%2fv1alpha1">gvisor.runtime.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="gvisor.runtime.extensions.config.gardener.cloud/v1alpha1">gvisor.runtime.extensions.config.gardener.cloud/v1alpha1</h2>
<p>

</p>

<h3 id="controllerconfiguration">ControllerConfiguration
</h3>


<p>
ControllerConfiguration defines the configuration for the GVisor runtime extension.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>clientConnection</code></br>
<em>
<a href="#clientconnectionconfiguration">ClientConnectionConfiguration</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClientConnection specifies the kubeconfig file and client connection<br />settings for the proxy server to use when communicating with the apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>healthCheckConfig</code></br>
<em>
<a href="#healthcheckconfig">HealthCheckConfig</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HealthCheckConfig is the config for the health check controller</p>
</td>
</tr>

</tbody>
</table>


<h3 id="gvisorconfiguration">GVisorConfiguration
</h3>


<p>
GVisorConfiguration defines the configuration for the gVisor runtime extension.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>configFlags</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConfigFlags is a map of additional flags that are passed to the runsc binary used by gVisor.</p>
</td>
</tr>

</tbody>
</table>


