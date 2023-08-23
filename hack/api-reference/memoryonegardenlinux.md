<p>Packages:</p>
<ul>
<li>
<a href="#memoryone-gardenlinux.os.extensions.gardener.cloud%2fv1alpha1">memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1">memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the v1alpha1 version of the API.</p>
</p>
Resource Types:
<ul><li>
<a href="#memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1.OperatingSystemConfiguration">OperatingSystemConfiguration</a>
</li></ul>
<h3 id="memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1.OperatingSystemConfiguration">OperatingSystemConfiguration
</h3>
<p>
<p>OperatingSystemConfiguration allows to specify configuration for the operating system.</p>
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
memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>OperatingSystemConfiguration</code></td>
</tr>
<tr>
<td>
<code>memoryTopology</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>MemoryTopology allows to configure the <code>mem_topology</code> parameter. If not present, it will default to <code>2</code>.</p>
</td>
</tr>
<tr>
<td>
<code>systemMemory</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SystemMemory allows to configure the <code>system_memory</code> parameter. If not present, it will default to <code>6x</code>.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
