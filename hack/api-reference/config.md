<p>Packages:</p>
<ul>
<li>
<a href="#gvisor.runtime.extensions.config.gardener.cloud%2fv1alpha1">gvisor.runtime.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="gvisor.runtime.extensions.config.gardener.cloud/v1alpha1">gvisor.runtime.extensions.config.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the GVisor container runtime configuration API resources.</p>
</p>
Resource Types:
<ul><li>
<a href="#gvisor.runtime.extensions.config.gardener.cloud/v1alpha1.ControllerConfiguration">ControllerConfiguration</a>
</li><li>
<a href="#gvisor.runtime.extensions.config.gardener.cloud/v1alpha1.GvisorConfiguration">GvisorConfiguration</a>
</li></ul>
<h3 id="gvisor.runtime.extensions.config.gardener.cloud/v1alpha1.ControllerConfiguration">ControllerConfiguration
</h3>
<p>
<p>ControllerConfiguration defines the configuration for the GVisor runtime extension.</p>
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
<code>apiVersion</code></br>
string</td>
<td>
<code>
gvisor.runtime.extensions.config.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>ControllerConfiguration</code></td>
</tr>
<tr>
<td>
<code>clientConnection</code></br>
<em>
<a href="https://godoc.org/k8s.io/component-base/config/v1alpha1#ClientConnectionConfiguration">
Kubernetes v1alpha1.ClientConnectionConfiguration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClientConnection specifies the kubeconfig file and client connection
settings for the proxy server to use when communicating with the apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>healthCheckConfig</code></br>
<em>
<a href="https://github.com/gardener/gardener/extensions/pkg/apis/config">
github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1.HealthCheckConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HealthCheckConfig is the config for the health check controller</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gvisor.runtime.extensions.config.gardener.cloud/v1alpha1.GvisorConfiguration">GvisorConfiguration
</h3>
<p>
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
<code>apiVersion</code></br>
string</td>
<td>
<code>
gvisor.runtime.extensions.config.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>GvisorConfiguration</code></td>
</tr>
<tr>
<td>
<code>additionalCapabilities</code></br>
<em>
<a href="#gvisor.runtime.extensions.config.gardener.cloud/v1alpha1.GVisorAdditionalCapabilities">
GVisorAdditionalCapabilities
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Capabilities to the granted to GVisor containers.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gvisor.runtime.extensions.config.gardener.cloud/v1alpha1.GVisorAdditionalCapabilities">GVisorAdditionalCapabilities
</h3>
<p>
(<em>Appears on:</em>
<a href="#gvisor.runtime.extensions.config.gardener.cloud/v1alpha1.GvisorConfiguration">GvisorConfiguration</a>)
</p>
<p>
<p>List of capabilities that can be granted to GVisor containers.</p>
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
<code>NET_RAW</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>SYS_ADMIN</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
