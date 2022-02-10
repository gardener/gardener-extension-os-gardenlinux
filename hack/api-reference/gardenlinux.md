<p>Packages:</p>
<ul>
<li>
<a href="#gardenlinux.os.extensions.gardener.cloud%2fv1alpha1">gardenlinux.os.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="gardenlinux.os.extensions.gardener.cloud/v1alpha1">gardenlinux.os.extensions.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the v1alpha1 version of the API.</p>
</p>
Resource Types:
<ul><li>
<a href="#gardenlinux.os.extensions.gardener.cloud/v1alpha1.OperatingSystemConfiguration">OperatingSystemConfiguration</a>
</li></ul>
<h3 id="gardenlinux.os.extensions.gardener.cloud/v1alpha1.OperatingSystemConfiguration">OperatingSystemConfiguration
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
gardenlinux.os.extensions.gardener.cloud/v1alpha1
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
<code>linuxSecurityModule</code></br>
<em>
<a href="#gardenlinux.os.extensions.gardener.cloud/v1alpha1.LinuxSecurityModule">
LinuxSecurityModule
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LinuxSecurityModule allows to configure default Linux Security Module for Garden Linux.</p>
</td>
</tr>
<tr>
<td>
<code>netfilterBackend</code></br>
<em>
<a href="#gardenlinux.os.extensions.gardener.cloud/v1alpha1.NetFilterBackend">
NetFilterBackend
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NetFilterBackend allows to configure the netfilter backend to be used on Garden Linux.</p>
</td>
</tr>
<tr>
<td>
<code>cgroupVersion</code></br>
<em>
<a href="#gardenlinux.os.extensions.gardener.cloud/v1alpha1.CgroupVersion">
CgroupVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CgroupVersion allows to configure which cgroup version will be used on Garden Linux</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gardenlinux.os.extensions.gardener.cloud/v1alpha1.CgroupVersion">CgroupVersion
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#gardenlinux.os.extensions.gardener.cloud/v1alpha1.OperatingSystemConfiguration">OperatingSystemConfiguration</a>)
</p>
<p>
<p>CgroupVersion defines the cgroup version (v1 or v2) to be configured on Garden Linux</p>
</p>
<h3 id="gardenlinux.os.extensions.gardener.cloud/v1alpha1.LinuxSecurityModule">LinuxSecurityModule
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#gardenlinux.os.extensions.gardener.cloud/v1alpha1.OperatingSystemConfiguration">OperatingSystemConfiguration</a>)
</p>
<p>
<p>LinuxSecurityModule defines the Linux Security Module (LSM) for Garden Linux</p>
</p>
<h3 id="gardenlinux.os.extensions.gardener.cloud/v1alpha1.NetFilterBackend">NetFilterBackend
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#gardenlinux.os.extensions.gardener.cloud/v1alpha1.OperatingSystemConfiguration">OperatingSystemConfiguration</a>)
</p>
<p>
<p>NetFilterBackend defines the netfilter backend for Garden Linux</p>
</p>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
