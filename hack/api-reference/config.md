<p>Packages:</p>
<ul>
<li>
<a href="#gvisor.runtime.extensions.config.gardener.cloud%2fv1alpha1">gvisor.runtime.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="gvisor.runtime.extensions.config.gardener.cloud/v1alpha1">gvisor.runtime.extensions.config.gardener.cloud/v1alpha1</h2>
<p>

</p>

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
<tr>
<td>
<code>testImageTag</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>TestImageTag is the tag for the gardener-extension-runtime-gvisor-installation image to be tested.<br />It requires that the `gvisorInstallation.testRepository` is configured in the operator extension values and<br />the image has been uploaded and tagged accordingly.<br />Only used for development and testing purposes. Does not work if `gvisorInstallation.testRepository` is not specified.</p>
</td>
</tr>

</tbody>
</table>


