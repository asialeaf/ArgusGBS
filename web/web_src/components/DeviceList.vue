<template>
	<div>
		<div class="box box-primary">
			<div class="box-header">
				<h4 class="text-primary text-center">设备列表</h4>
			</div>
			<div class="box-body">
				<div class="form-inline" autocomplete="off" spellcheck="false">
					<div class="form-group form-group-sm">
						<label>搜索</label>
						<div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                            <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出设备列表"><i class="fa fa-download"></i></button>
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['uploadDlg'].show()" title="导入设备列表"><i class="fa fa-upload"></i></button>
                            </div>
						</div>
						<input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
					</div>
					<span class="hidden-xs">&nbsp;&nbsp;</span>
					<div class="form-group form-group-sm">
						<label>设备类型</label>
						<select class="form-control" v-model.trim="type">
							<option value="">全部</option>
							<option value="131,132">摄像机</option>
							<option value="111,118,130,210">录像机</option>
							<option value="114">解码器</option>
							<option value="200">下级平台</option>
						</select>
					</div>
					<span class="hidden-xs">&nbsp;&nbsp;</span>
					<div class="form-group form-group-sm">
						<label>在线状态</label>
						<select class="form-control" v-model.trim="online">
							<option value="">全部</option>
							<option value="true">在线</option>
							<option value="false">离线</option>
						</select>
					</div>
					<!-- <div class="form-group pull-right">
						<router-link :to="`/devices/tree`" class="btn btn-default btn-sm">
								<i class="fa fa-sitemap"></i> 树视图
						</router-link>
					</div> -->
				</div>
				<br>
				<div class="clearfix"></div>
				<el-table :data="devices" stripe :default-sort="{prop: 'ID', order: 'ascending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
					<el-table-column prop="ID" label="设备国标编号" min-width="200" sortable="custom" show-overflow-tooltip></el-table-column>
					<el-table-column label="操作" :min-width="hasAnyRole(serverInfo, userInfo, '管理员', '操作员') ? 240 : 120" :fixed="isMobile() ? false : 'right'">
						<template slot-scope="props">
                            <div class="btn-group btn-group-xs">
                                <router-link class="btn btn-info" :to="`/devices/channels/${props.row.ID}/1`">
                                    <i class="fa fa-info"></i> 查看通道
                                </router-link>
                                <button type="button" class="btn btn-primary" @click.prevent="fetchCatalog(props.row)" v-if="props.row.Online && hasAnyRole(serverInfo, userInfo, '管理员', '操作员')"
                                    :disabled="(catalogMap[props.row.ID] || props.row.CatalogProgress) && !debug" :title="formatCatalogTitle(props.row)">
                                    <i :class="['fa', 'fa-refresh', {'fa-spin': catalogMap[props.row.ID] || props.row.CatalogProgress}]"></i> 更新通道
                                </button>
                                <button type="button" class="btn btn-warning" @click.prevent="editDevice(props.row)" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                                    <i class="fa fa-edit"></i> 编辑
                                </button>
                                <button type="button" class="btn btn-danger" @click.prevent="removeDevice(props.row)" v-if="!props.row.Online && hasAnyRole(serverInfo, userInfo, '管理员')">
                                    <i class="fa fa-remove"></i> 删除
                                </button>
                            </div>
						</template>
					</el-table-column>
					<el-table-column prop="Name" label="名称" min-width="140" show-overflow-tooltip>
						<template slot-scope="props">
							<span :class="{'text-orange': !!props.row.CustomName}" :title="props.row.CustomName ? props.row.Name||'-' : ''">
								{{props.row.CustomName||props.row.Name||'-'}}
							</span>
						</template>
					</el-table-column>
					<!-- <el-table-column prop="CommandTransport" label="信令传输" min-width="100" :formatter="formatName" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" show-overflow-tooltip></el-table-column> -->
					<el-table-column prop="MediaTransport" label="流传输模式" min-width="140" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
						<template slot-scope="props">
							<el-dropdown size="small" trigger="click" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" @command="setMediaTransport">
								<span class="el-dropdown-link">
									{{formatTransport(props.row)}} <i class="el-icon-arrow-down el-icon--right"></i>
								</span>
								<el-dropdown-menu slot="dropdown">
									<el-dropdown-item :command="{row: props.row, MediaTransport: 'UDP', MediaTransportMode: 'passive'}">UDP</el-dropdown-item>
									<el-dropdown-item :command="{row: props.row, MediaTransport: 'TCP', MediaTransportMode: 'passive'}">TCP 被动</el-dropdown-item>
									<el-dropdown-item :command="{row: props.row, MediaTransport: 'TCP', MediaTransportMode: 'active'}">TCP 主动</el-dropdown-item>
								</el-dropdown-menu>
							</el-dropdown>
							<span v-else>{{formatTransport(props.row)}}</span>
						</template>
					</el-table-column>
					<el-table-column prop="ChannelCount" label="通道数" min-width="120" sortable="custom" v-if="hasAnyRole(serverInfo, userInfo, '管理员') && hasAllChannel(serverInfo, userInfo)" show-overflow-tooltip>
						<template slot-scope="props">
							<span class="text-red" v-if="props.row.ChannelOverLoad && props.row.Online">
								{{props.row.ChannelCount}} (授权满)
							</span>
							<span v-else-if="props.row.CatalogProgress && props.row.CatalogProgress.split('/').length == 2">{{props.row.ChannelCount}}
								<el-progress :percentage="formatCatalogPercent(props.row)" text-inside :stroke-width="14" :title="formatCatalogTitle(props.row)"></el-progress>
							</span>
							<span v-else>{{props.row.ChannelCount}}</span>
						</template>
					</el-table-column>
					<el-table-column prop="Online" label="在线状态" min-width="100">
						<template slot-scope="props">
							<span v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
								<a role="button" @click.prevent="showStatusLog(props.row)"
									:class="[{'text-success': props.row.Online, 'text-gray': !props.row.Online}]" title="点击查看设备状态记录">
									{{props.row.Online ? "在线" : "离线"}}
								</a>
								<a role="button" @click.prevent="showStreamLog(props.row)"
									:class="[{'text-success': props.row.Online, 'text-gray': !props.row.Online}]" title="点击查看设备流量统计">
									<i class="fa fa-bar-chart"></i>
								</a>
							</span>
							<span v-else-if="props.row.Online" class="text-success">
								在线
							</span>
							<span v-else class="text-gray">
								离线
							</span>
						</template>
					</el-table-column>
					<el-table-column prop="RemoteIP" label="出口 IP" min-width="210" :formatter="formatRemoteIP" v-if="hasAnyRole(serverInfo, userInfo, '管理员') && network" show-overflow-tooltip></el-table-column>
					<!-- <el-table-column prop="RemotePort" label="端口" min-width="80" v-if="hasAnyRole(serverInfo, userInfo, '管理员')"></el-table-column> -->
					<el-table-column prop="RemoteRegion" label="所在地区" min-width="150" v-if="hasAnyRole(serverInfo, userInfo, '管理员') && region" show-overflow-tooltip></el-table-column>
					<el-table-column prop="Manufacturer" label="厂家" min-width="150" show-overflow-tooltip>
						<template slot-scope="props">
							<span>{{props.row.Manufacturer || '-'}}</span><span v-if="props.row.GBVer">, GB{{props.row.GBVer}}</span>
							<span class="badge" v-if="props.row.Type == 'Decode'">解码器</span>
						</template>
					</el-table-column>
					<el-table-column prop="LastKeepaliveAt" label="最近心跳" min-width="160" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" sortable="custom"></el-table-column>
					<el-table-column prop="LastRegisterAt" label="最近注册" min-width="160" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" sortable="custom"></el-table-column>
					<el-table-column prop="UpdatedAt" label="更新时间" min-width="160" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" sortable="custom"></el-table-column>
					<el-table-column prop="CreatedAt" label="创建时间" min-width="160" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" sortable="custom"></el-table-column>
				</el-table>
			</div>
			<div class="box-footer" v-if="total > 0">
				<el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
			</div>
		</div>
		<div class="alert text-center" v-if="serverInfo.DemoUser">
			<small>
                <strong><i class="fa fa-info-circle"></i> 提示 : </strong>
                可向 Redis 订阅 device 消息以获取设备实时状态 > SUBSCRIBE device; 消息内容为 "设备国标编号 ON/OFF"
			</small>
		</div>
		<UploadDlg ref="uploadDlg" title="上传设备列表" url="/api/v1/device/import" @uploaded="uploaded"></UploadDlg>
		<DeviceEditDlg ref="deviceEditDlg" @submit="getDeviceList" :useSeparateDevicePassword="serverInfo.UseSeparateDevicePassword === true"></DeviceEditDlg>
		<DeviceLogDlg ref="deviceLogDlg" :serverInfo="serverInfo" :userInfo="userInfo"></DeviceLogDlg>
		<el-dialog title="提示" :visible.sync="removeDlg" :lock-scroll="false">
			<p>确认删除 {{removeName || removeID}} ?</p>
			<br>
			<div class="checkbox" v-if="removeUA">
				<el-checkbox style="margin-left:-19px;margin-top:-5px;" v-model.trim="removeUACheck">
                    批量删除 <strong>{{removeUA}}</strong>
				</el-checkbox>
				<br><br>
				<el-checkbox style="margin-left:-19px;margin-top:-5px;" v-model.trim="forbidUACheck" v-show="removeUACheck">
                    禁止 <strong>{{removeUA}}</strong> 接入
				</el-checkbox>
			</div>
			<div class="checkbox" v-else-if="removeIP">
				<el-checkbox style="margin-left:-19px;margin-top:-5px;" v-model.trim="removeIPCheck">
                    批量删除 <strong>{{removeIP}}</strong>
				</el-checkbox>
				<br><br>
				<el-checkbox style="margin-left:-19px;margin-top:-5px;" v-model.trim="forbidIPCheck" v-show="removeIPCheck">
                    禁止 <strong>{{removeIP}}</strong> 接入
				</el-checkbox>
			</div>
			<span slot="footer" class="dialog-footer">
				<el-button @click="removeDlg = false" size="small">取 消</el-button>
				<el-button type="primary" @click="removeBatch" size="small">确 定</el-button>
			</span>
		</el-dialog>
	</div>
</template>

<script>
import _ from "lodash";
import UploadDlg from "components/UploadDlg.vue";
import DeviceEditDlg from 'components/DeviceEditDlg.vue';
import DeviceLogDlg from "components/DeviceLogDlg";
import { mapState } from "vuex";
export default {
	props: {},
	data() {
		return {
			q: "",
			type: "",
			online: "",
			total: 0,
			pageSize: 10,
			currentPage: 1,
			sort: "ID",
			order: "asc",
			devices: [],
			loading: false,
			bgLoading: false,
			timer: 0,
			statusLogs:[],
			bStatusLogLoading: false,
			removeDlg: false,
			removeUACheck: false,
			removeIPCheck: false,
			forbidUACheck: false,
			forbidIPCheck: false,
			removeID: "",
			removeName: "",
			removeUA: "",
			removeIP: "",
			debug: false,
			checkCatalogProgress: true,
			catalogMap: {},
			network: false,
			region: false,
		};
	},
	components: {
		UploadDlg, DeviceEditDlg, DeviceLogDlg
	},
	computed: {
		...mapState(["userInfo", "serverInfo"])
	},
	mounted() {
		// this.$refs["q"].focus();
		// this.getDeviceList();
		// this.timer = setInterval(() => {
		//     this.getDeviceList();
		// }, 3000);
		$(document).on("keydown", this.keyDown);
	},
	beforeDestroy() {
		if (this.timer) {
			clearInterval(this.timer);
			this.timer = 0;
		}
		$(document).off("keydown", this.keyDown);
	},
	methods: {
		ready() {
			this.$watch('q', function(newVal, oldVal) {
				this.doDelaySearch();
			});
			this.$watch('online', function(newVal, oldVal) {
				this.doSearch();
			});
			this.$watch('type', function(newVal, oldVal) {
				this.doSearch();
			});
			this.$watch('currentPage', function(newVal, oldVal) {
				this.doSearch(newVal);
			});
			// this.getDeviceList();
			this.timer = setInterval(() => {
				this.getDeviceList(false);
			}, 3000);
		},
		doSearch(page = 1) {
			var query = {};
			if(this.q) query["q"] = this.q;
			if(this.type) query["type"] = this.type;
			if(this.online) query["online"] = this.online;
			this.$router.replace({
				path: `/devices/${page}`,
				query: query
			});
		},
		doDelaySearch: _.debounce(function() {
			this.doSearch();
		}, 800),
		getDeviceList(global = true) {
			if(global) {
				this.loading = true;
			} else {
				if(this.bgLoading || this.loading) return;
				this.bgLoading = true;
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/device/list",
				global: global,
				data: {
					q: this.q,
					start: (this.currentPage -1) * this.pageSize,
					limit: this.pageSize,
					device_type: this.type,
					online: this.online,
					sort: this.sort,
					order: this.order,
					check_catalog_progress: this.checkCatalogProgress,
				}
			}).then(ret => {
				this.total = ret["DeviceCount"];
				this.devices = ret["DeviceList"];
				this.network = ret["DeviceNetwork"] !== false;
				this.region = !!ret["DeviceRegion"];
			}).always(() => {
				if(global) {
					this.loading = false;
				} else {
					this.bgLoading = false;
				}
			});
		},
		sortChange(data) {
			this.sort = data.prop;
			this.order = data.order == "ascending" ? "asc" : "desc";
			this.getDeviceList();
		},
		fetchCatalog(row) {
			if(this.serverInfo && !this.serverInfo.RemainDays) {
				this.$message({
					type: "error",
					message: "授权过期"
				})
				return
			}
			this.$set(this.catalogMap, row.ID, true);
			$.ajax({
				method: "GET", // type: "GET", is also ok, method since jquery 1.9
				url: "/api/v1/device/fetchcatalog",
				global: false,
				data: {
					serial: row.ID
				}
			}).then(ret => {
				this.$message({
					type: "success",
					message: "更新通道信息成功"
				})
			}).fail(xhr => {
				xhr && console.log(`fetch catalog ajax error: ${xhr.status} ${xhr.responseText}`);
			}).always(() => {
				this.$delete(this.catalogMap, row.ID);
			})
		},
		download() {
			window.open(`/api/v1/device/export?q=${this.q}&device_type=${this.type}&online=${this.online}`);
		},
		uploaded() {
            this.$message({
                type: 'success',
                message: "上传成功！"
            })
            this.getDeviceList();
            this.$refs['uploadDlg'].hide();
		},
		removeBatch(){
			this.removeDlg = false;
			var data = {
				serial: this.removeID,
			};
			if (this.removeUACheck && this.removeUA) {
				data["ua"] = this.removeUA;
				if (this.forbidUACheck) {
					data["forbid"] = true;
				}
			} else if (this.removeIPCheck && this.removeIP) {
				data["ip"] = this.removeIP;
				if (this.forbidIPCheck) {
					data["forbid"] = true;
				}
			}
			$.post("/api/v1/device/remove", data).always(() => {
				this.getDeviceList();
			})
		},
		removeDevice(row) {
			if (!row.Online && !row.ChannelCount && row.LastKeepaliveAt === "0001-01-01 00:00:00" && (row.Manufacturer || row.RemoteIP) && row.ID && row.ID.length < 10) {
			// if (!row.Online && !row.ChannelCount && row.LastKeepaliveAt === "0001-01-01 00:00:00" && row.Manufacturer && row.ID) {
				this.removeID = row.ID;
				this.removeName = row.Name;
				this.removeUA = row.Manufacturer;
				this.removeIP = row.RemoteIP;
				this.removeUACheck = false;
				this.removeIPCheck = false;
				this.forbidUACheck = false;
				this.forbidIPCheck = false;
				this.removeDlg = true;
				return
			}
			this.$confirm(`确认删除 ${row.Name || row.ID} ?`, "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/device/remove", {
					serial: row.ID
				}).always(() => {
					this.getDeviceList();
				});
			}).catch(() => {});
		},
		editDevice(row) {
			this.$refs["deviceEditDlg"].show({
				serial: row.ID,
				name: row.Name,
				custom_name: row.CustomName,
				media_transport: row.MediaTransport,
				media_transport_mode: row.MediaTransportMode,
				stream_mode: row.StreamMode,
				recv_stream_ip: row.RecvStreamIP,
				contact_ip: row.ContactIP,
				sms_id: row.SMSID,
				sms_group_id: row.SMSGroupID,
				charset: row.Charset,
				catalog_interval: row.CatalogInterval,
				subscribe_interval: row.SubscribeInterval,
				catalog_subscribe: row.CatalogSubscribe,
				alarm_subscribe: row.AlarmSubscribe,
				position_subscribe: row.PositionSubscribe,
				ptz_subscribe: row.PTZSubscribe,
				password: row.Password,
				record_center: row.RecordCenter,
				record_indistinct: row.RecordIndistinct,
				civil_code_first: row.CivilCodeFirst,
				keep_original_tree: row.KeepOriginalTree,
				drop_channel_type: row.DropChannelType,
				longitude: row.Longitude,
				latitude: row.Latitude,
			});
		},
		formatName(row, col, cell) {
			if(cell) return cell;
			return "-";
		},
		formatRemoteIP(row, col, cell) {
			if(!row.RemoteIP) return "-";
			var ret = row.RemoteIP;
			if(row.RemotePort) {
				ret = `${row.RemoteIP}:${row.RemotePort}`;
			}
			if(row.CommandTransport) {
				ret = `${String(row.CommandTransport).toLowerCase()}://${ret}`;
			}
			return ret;
		},
		formatTransport(row, col, cell) {
			var ret = String(row.MediaTransport).toUpperCase();
			if(ret == "TCP") {
				ret += row.MediaTransportMode == 'active' ? " 主动" : " 被动";
			}
			return ret;
		},
		formatCatalogTitle(row) {
			if(row.CatalogProgress) return `更新中(${row.CatalogProgress})`;
			if(this.catalogMap[row.ID]) return "更新中...";
			return "";
		},
		formatCatalogPercent(row) {
			if(!row || !row.CatalogProgress) return 0;
			var parts = row.CatalogProgress.split('/');
			if(parts.length != 2) return 0;
			var den = parseInt(parts[1], 10);
			if(!den) return 0;
			var num = parseInt(parts[0], 10) || 0;
			var p = parseInt(num * 100 / den, 10);
			if(isNaN(p) || p < 0) return 0;
			if(p > 100) return 100;
			return p;
		},
		setMediaTransport(cmd) {
			var transport = cmd.MediaTransport || cmd.row.MediaTransport;
			var transportMode = cmd.MediaTransportMode || cmd.row.MediaTransportMode;
			$.post("/api/v1/device/setmediatransport", {
				serial: cmd.row.ID,
				media_transport: transport,
				media_transport_mode: transportMode,
			}).then(() => {
				this.$set(cmd.row, "MediaTransport", transport);
				this.$set(cmd.row, "MediaTransportMode", transportMode);
			})
		},
		showDeviceLog(row) {
			this.$refs["deviceLogDlg"].show(`设备(${row.ID})`, row.ID);
		},
		showStatusLog(row) {
			this.$refs["deviceLogDlg"].showStatusLog(`设备(${row.ID})`, row.ID);
		},
		showStreamLog(row) {
			this.$refs["deviceLogDlg"].showStreamLog(`设备(${row.ID})`, row.ID);
		},
		keyDown(e) {
			if(e.altKey && e.shiftKey) {
				switch(e.key) {
					case 'D':
						e.preventDefault();
						this.toggleDebug();
						break;
				}
			}
		},
		toggleDebug() {
			this.debug = !this.debug;
		},
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.q = to.query.q || "";
			vm.type = to.query.type || "";
			vm.online = to.query.online || "";
			vm.checkCatalogProgress = (to.query.checkCatalogProgress || "yes") === "yes";
			vm.currentPage = parseInt(to.params.page) || 1;
			vm.ready();
		});
	},
	beforeRouteLeave(to, from, next) {
		if (this.timer) {
			clearInterval(this.timer);
			this.timer = 0;
		}
		next();
	},
	beforeRouteUpdate(to, from, next) {
		next();
		this.$nextTick(() => {
			this.q = to.query.q || "";
			this.type = to.query.type || "";
			this.online = to.query.online || "";
			this.checkCatalogProgress = (to.query.checkCatalogProgress || "yes") === "yes";
			this.currentPage = parseInt(to.params.page) || 1;
			this.devices = [];
			this.total = 0;
			this.getDeviceList();
		});
	}
};
</script>
