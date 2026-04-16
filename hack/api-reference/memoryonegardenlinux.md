<p>Packages:</p>
<ul>
<li>
<a href="#memoryone-gardenlinux.os.extensions.gardener.cloud%2fv1alpha1">memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1">memoryone-gardenlinux.os.extensions.gardener.cloud/v1alpha1</h2>
<p>

</p>

<h3 id="operatingsystemconfiguration">OperatingSystemConfiguration
</h3>


<p>
OperatingSystemConfiguration allows to specify configuration for the operating system.
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
<code>memoryTopology</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>MemoryTopology allows to configure the `mem_topology` parameter. If not present, it will default to `2`.</p>
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
<p>SystemMemory allows to configure the `system_memory` parameter. If not present, it will default to `6x`.</p>
</td>
</tr>

</tbody>
</table>


