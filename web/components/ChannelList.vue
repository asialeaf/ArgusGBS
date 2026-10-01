<template>
<div>
    <div class="box box-primary">
        <div class="box-header">
            <h4 class="text-primary text-center">通道列表({{name||devid}}<span v-if="dir_serial">:{{dir_serial}}</span>)</h4>
        </div>
        <div class="box-body">
            <div class="form-inline" autocomplete="off" spellcheck="false">
                <div class="form-group form-group-sm" v-if="$getQueryString('menu') != 'no'">
                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="$router.go(-1)">
                        <i class="fa fa-chevron-left"></i> 返回
                    </button>
                </div>
                <span class="hidden-xs">&nbsp;&nbsp;</span>
                <div class="form-group form-group-sm">
                    <label>搜索</label>
                    <div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                        <div class="input-group-btn">
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['downloadDlg'].show(devid)" title="导出通道列表"><i class="fa fa-download"></i></button>
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['uploadDlg'].show()" title="导入通道列表"><i class="fa fa-upload"></i></button>
                        </div>
                    </div>
                    <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                </div>
                <span class="hidden-xs">&nbsp;&nbsp;</span>
                <div class="form-group form-group-sm">
                    <label>通道类型</label>
                    <select class="form-control" v-model.trim="type">
                        <option value="">全部</option>
                        <option value="device">设备</option>
                        <option value="dir">子目录</option>
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
            </div>
            <br>
            <el-table :data="channels" stripe :default-sort="{prop: 'Channel', order: 'ascending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                <el-table-column prop="Channel" label="通道号" min-width="100" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column label="操作" min-width="240" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                    <template slot-scope="props">
                        <div class="btn-group btn-group-xs">
                            <button type="button" class="btn btn-primary" @click.prevent="playStream(props.row)" :disabled="props.row.Locked" v-if="canPlay(props.row)">
                                <i class="fa fa-play-circle"></i> 播放
                            </button>
                            <el-dropdown class="btn btn-primary" style="padding-left: 1px;" :show-timeout="0" v-if="canPlay(props.row) && canShare(props.row)">
                                <i class="el-icon-arrow-down el-icon--right" style="color: white;"></i>
                                <el-dropdown-menu slot="dropdown">
									<a :href="`/play.html?serial=${props.row.DeviceID}&code=${props.row.ID}`" target="_blank">
										<el-dropdown-item>
											<i class="fa fa-share"></i> 分享页
										</el-dropdown-item>
									</a>
                                </el-dropdown-menu>
                            </el-dropdown>
                            <button type="button" class="btn btn-primary" @click.prevent="decodeStart(props.row)" :disabled="props.row.Locked" v-if="canDecode(props.row)">
                                <i class="fa fa-play-circle"></i> 解码
                            </button>
                            <button type="button" class="btn btn-danger" @click.prevent="stopStream(props.row)" v-if="props.row.StreamID && hasAnyRole(serverInfo, userInfo, '管理员')">
                                <i class="fa fa-stop"></i> 停止
                            </button>
                            <router-link class="btn btn-info" :to="`/devices/playback/timebox/${devid}/${props.row.ID}`" v-if="canPlayback(props.row)">
                                <i class="fa fa-file-video-o"></i> 设备录像
                            </router-link>
                            <el-dropdown class="btn btn-info" style="padding-left: 1px;" :show-timeout="0" v-if="canPlayback(props.row) && canCloudRecord(props.row)">
                                <i class="el-icon-arrow-down el-icon--right" style="color: white;"></i>
                                <el-dropdown-menu slot="dropdown">
									<router-link :to="`/cloudrecord/timeview/${devid}/${props.row.ID}`">
										<el-dropdown-item>
											<i class="fa fa-file-video-o"></i> 云端录像
										</el-dropdown-item>
									</router-link>
                                </el-dropdown-menu>
                            </el-dropdown>
                            <router-link class="btn btn-info" :to="`/cloudrecord/timeview/${devid}/${props.row.ID}`" v-else-if="!props.row.DeviceOnline && canCloudRecord(props.row)">
                                <i class="fa fa-file-video-o"></i> 云端录像
                            </router-link>
                            <!-- <router-link class="btn btn-info" :to="`/devices/channels/${devid}/1?dir_serial=${props.row.ID}`" v-if="props.row.SubCount > 0"> -->
                            <router-link class="btn btn-info" :to="`/devices/channels/${devid}/1?dir_serial=${props.row.ID}`" v-if="isDir(props.row)" exact>
                                <i class="fa fa-info"></i> 查看子目录
                            </router-link>
                            <button type="button" class="btn btn-primary" @click.prevent="fetchCatalog(props.row)" v-if="canFetchCatalog(props.row) && hasAnyRole(serverInfo, userInfo, '管理员')"
                                :disabled="catalogMap[`${devid}:${props.row.ID}`]" :title="formatCatalogTitle(props.row)">
                                <i :class="['fa', 'fa-refresh', {'fa-spin': catalogMap[`${devid}:${props.row.ID}`]}]"></i> 更新通道
                            </button>
                            <button type="button" class="btn btn-danger" @click.prevent="removeVirtual(props.row)" v-if="props.row.Manufacturer == 'virtual' && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                                <i class="fa fa-remove"></i> 删除
                            </button>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="ID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column prop="Name" label="通道名称" min-width="120" show-overflow-tooltip>
                    <template slot-scope="props">
                        <a role="button" @click.prevent="editChannelName(props.row)" :class="{'text-orange': !!props.row.CustomName}" :title="props.row.CustomName ? props.row.Name||'-' : ''"
                            v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                            {{props.row.CustomName||props.row.Name||'-'}}
                        </a>
                        <span v-else>
                            {{props.row.CustomName||props.row.Name||'-'}}
                        </span>
                    </template>
                </el-table-column>
                <el-table-column min-width="100" label="快照">
                    <template slot-scope="props">
                        <span v-if="isDir(props.row)">
                            <i class="fa fa-sitemap"></i>
                        </span>
                        <el-popover v-else :open-delay="1000" :close-delay="10" placement="left" :title="props.row.ID" width="400" trigger="hover">
                            <img onerror="this.src='/images/default_snap.png';" style="width:100%;height:100%;" :src="props.row.SnapURL">
                            <img onerror="this.src='/images/default_snap.png';" style="height:30px;width:50px;" slot="reference" :src="props.row.SnapURL">
                        </el-popover>
                    </template>
                </el-table-column>
                <el-table-column prop="Status" label="在线状态" min-width="100">
                    <template slot-scope="props">
                        <!-- <span v-if="props.row.SubCount > 0">-</span> -->
                        <span v-if="isDir(props.row)">-</span>
                        <span v-else-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                            <a role="button" @click.prevent="showStatusLog(props.row)" title="点击查看通道状态记录"
                                :class="[{'text-success': props.row.Status == 'ON', 'text-orange': !!props.row.CustomStatus, 'text-gray': (props.row.CustomStatus || props.row.Status) != 'ON'}]">
                                {{(props.row.CustomStatus || props.row.Status) == 'ON' ? "在线" : "离线"}}
                            </a>
                            <a role="button" @click.prevent="showStreamLog(props.row)" title="点击查看通道流量统计"
                                :class="{'text-success': props.row.Status == 'ON', 'text-orange': !!props.row.CustomStatus, 'text-gray': (props.row.CustomStatus || props.row.Status) != 'ON'}">
                                <i class="fa fa-bar-chart"></i>
                            </a>
                        </span>
                        <span v-else-if="(props.row.CustomStatus || props.row.Status) == 'ON'" :class="[{'text-success': props.row.Status == 'ON', 'text-orange': !!props.row.CustomStatus}]">
                            在线
                        </span>
                        <span v-else class="text-gray">
                            离线
                        </span>
                    </template>
                </el-table-column>
                <el-table-column prop="AudioEnable" label="开启音频" min-width="100" v-if="!serverInfo.GlobalChannelAudio && hasAnyRole(serverInfo, userInfo, '管理员')">
                    <template slot-scope="props">
                        <el-switch :value="props.row.AudioEnable" @change="toggleAudio(props.row)" :disabled="isDir(props.row) || props.row.DeviceType == 'Decode'"></el-switch>
                    </template>
                </el-table-column>
                <el-table-column prop="Ondemand" label="按需直播" min-width="100" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                    <template slot-scope="props">
                        <el-switch :value="props.row.Ondemand" @change="toggleOndemand(props.row)" :disabled="isDir(props.row) || props.row.DeviceType == 'Decode'"></el-switch>
                    </template>
                </el-table-column>
                <el-table-column prop="CloudRecord" label="云端录像" min-width="100" v-if="hasAnyRole(serverInfo, userInfo, '管理员') && serverInfo.VersionType != '标准版'">
                    <template slot-scope="props">
                        <el-switch :value="props.row.CloudRecord" @change="toggleCloudRecord(props.row)" :disabled="isDir(props.row) || props.row.DeviceType == 'Decode'"></el-switch>
                    </template>
                </el-table-column>
                <el-table-column prop="Shared" label="开启分享" min-width="100" v-if="!serverInfo.GlobalChannelShared && hasAnyRole(serverInfo, userInfo, '管理员')">
                    <template slot-scope="props">
                        <el-switch :value="props.row.Shared" @change="toggleShared(props.row)" :disabled="isDir(props.row) || props.row.DeviceType == 'Decode'"></el-switch>
                    </template>
                </el-table-column>
                <el-table-column prop="NumOutputs" label="在线人数" min-width="100" v-if="hasAnyRole(serverInfo, userInfo, '管理员')"></el-table-column>
                <el-table-column prop="SubCount" label="子节点数" min-width="100" v-if="hasAnyRole(serverInfo, userInfo, '管理员') && hasAllChannel(serverInfo, userInfo)"></el-table-column>
                <el-table-column prop="Manufacturer" label="厂家" min-width="120" show-overflow-tooltip>
                    <template slot-scope="props">
                        <span :class="{'text-orange': !!props.row.CustomManufacturer}">
                            {{props.row.CustomManufacturer||props.row.Manufacturer||'-'}}
                        </span>
                    </template>
                </el-table-column>
                <el-table-column prop="PTZType" label="云台类型" min-width="140" v-if="hasAnyRole(serverInfo, userInfo, '管理员', '操作员')">
                    <template slot-scope="props">
                        <el-dropdown size="small" trigger="click" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" @command="setPTZType">
                            <span :class="['el-dropdown-link', {'text-orange': props.row.CustomPTZType}]">
                                {{formatPTZType(props.row)}} <i class="el-icon-arrow-down el-icon--right"></i>
                            </span>
                            <el-dropdown-menu slot="dropdown">
                                <el-dropdown-item :command="{row: props.row, PTZType: 0}">0-未知</el-dropdown-item>
                                <el-dropdown-item :command="{row: props.row, PTZType: 1}">1-球机</el-dropdown-item>
                                <el-dropdown-item :command="{row: props.row, PTZType: 2}">2-半球机</el-dropdown-item>
                                <el-dropdown-item :command="{row: props.row, PTZType: 3}">3-固定枪机</el-dropdown-item>
                                <el-dropdown-item :command="{row: props.row, PTZType: 4}">4-遥控枪机</el-dropdown-item>
                            </el-dropdown-menu>
                        </el-dropdown>
                        <span :class="{'text-orange': props.row.CustumPTZType}" v-else>{{formatPTZType(props.row)}}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="Longitude" label="位置" min-width="180" show-overflow-tooltip>
                    <template slot-scope="props">
                        <a role="button" @click.prevent="editChannelPosition(props.row)" :class="{'text-orange': (props.row.CustomLongitude || props.row.CustomLatitude)}" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                            {{formatPosition(props.row)}}
                        </a>
                        <span v-else :class="{'text-orange': (props.row.CustomLongitude || props.row.CustomLatitude)}">{{formatPosition(props.row)}}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="UpdatedAt" label="更新时间" min-width="160" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" sortable="custom"></el-table-column>
                <el-table-column prop="CreatedAt" label="创建时间" min-width="160" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" sortable="custom"></el-table-column>
            </el-table>
        </div>
        <div class="box-footer" v-if="total > 0">
			<span class="text-danger" style="font-size:13px;line-height:32px;" v-if="!serverInfo.DemoUser && (serverInfo.GlobalChannelShared === 1 || (!serverInfo.GlobalChannelShared && channels.some(item => item.Shared)))">
				<strong><i class="fa fa-info-circle"></i> 提示 : </strong>
				<span v-if="serverInfo.GlobalChannelShared === 1">
					基础配置已开启全局通道分享，
				</span>
				<span v-else>
					当前列表包含开启分享通道，
				</span>
				视频可在公开未登录的情况下被他人访问，请注意数据安全！
			</span>
            <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
        </div>
    </div>
    <ChannelExportDlg ref="downloadDlg" :q="q" :online="online" :dir_serial="dir_serial" :channel_type="type"/>
    <UploadDlg ref="uploadDlg" title="上传通道列表" :url="uploadURL" @uploaded="uploaded" @hide="uploadThisDev = false">
        <template slot="header">
            <div class="checkbox text-center" v-if="devid">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" v-model.trim="uploadThisDev">
                    只导入当前设备({{devid}})的通道
                </el-checkbox>
                <br>
                <br>
            </div>
        </template>
    </UploadDlg>
    <ChannelNameEditDlg ref="channelNameEditDlg" :serverInfo="serverInfo" :userInfo="userInfo" @submit="getChannels"></ChannelNameEditDlg>
    <ChannelPositionEditDlg ref="channelPositionEditDlg" @submit="getChannels"></ChannelPositionEditDlg>
    <VideoDlg ref="videoDlg" :serverInfo="serverInfo" :userInfo="userInfo"></VideoDlg>
    <DeviceLogDlg ref="deviceLogDlg" :serverInfo="serverInfo" :userInfo="userInfo"></DeviceLogDlg>
	<DecodeChannelListDlg ref="decodeChannelListDlg" :serverInfo="serverInfo" :userInfo="userInfo" @selected="decodeStream"></DecodeChannelListDlg>
</div>
</template>

<script>
import _ from "lodash";
import ChannelExportDlg from "components/ChannelExportDlg.vue";
import UploadDlg from "components/UploadDlg.vue";
import ChannelNameEditDlg from "components/ChannelNameEditDlg.vue";
import ChannelPositionEditDlg from "components/ChannelPositionEditDlg.vue";
import VideoDlg from "components/VideoDlg.vue";
import DeviceLogDlg from "components/DeviceLogDlg.vue";
import DecodeChannelListDlg from "components/DecodeChannelListDlg.vue";
import { mapState } from "vuex";
export default {
	props: {
		devid: {
			type: String,
			default: ""
		},
		name: {
			type: String,
			default: ""
		}
	},
	data() {
		return {
			q: "",
			type: "",
			online: "",
			devOnline: false,
			dir_serial: "",
			total: 0,
			pageSize: 10,
			currentPage: 1,
			sort: "Channel",
			order: "asc",
			loading: false,
			bgLoading: false,
			xhr: null,
			timer: 0,
			channels: [],
			uploadThisDev: false,
			catalogMap: {},
			sharedSecurityTipShown: false,
		};
	},
	computed: {
		...mapState(['userInfo', 'serverInfo']),
		uploadURL() {
			if (this.uploadThisDev && this.devid) {
				return `/api/v1/channel/import?serial=${this.devid}`;
			}
			return "/api/v1/channel/import";
		}
	},
	// watch: {
	//   q: function(newVal, oldVal) {
	//     this.doDelaySearch();
	//   },
	//   online: function(newVal, oldVal) {
	//     this.doSearch();
	//   },
	//   type: function(newVal, oldVal) {
	//     this.doSearch();
	//   },
	//   dir_serial: function(newVal, oldVal) {
	//     this.doSearch();
	//   },
	//   currentPage: function(newVal, oldVal) {
	//     this.doSearch(newVal);
	//   }
	// },
	components: {
		ChannelExportDlg, UploadDlg, ChannelNameEditDlg, ChannelPositionEditDlg, VideoDlg, DeviceLogDlg, DecodeChannelListDlg
	},
	mounted() {
		// this.$refs["q"].focus();
		// this.getChannels();
		// this.timer = setInterval(() => {
		//     this.getChannels();
		// }, 3000);
	},
	beforeDestroy() {
		if (this.timer) {
			clearInterval(this.timer);
			this.timer = 0;
		}
	},
	methods: {
		ready() {
			this.unWatchQ = this.$watch('q', function(newVal, oldVal) {
				this.doDelaySearch();
			});
			this.unWatchOnline = this.$watch('online', function(newVal, oldVal) {
				this.doSearch();
			});
			this.unWatchChannelType = this.$watch('type', function(newVal, oldVal) {
				this.doSearch();
			});
			this.unWatchDirSerial = this.$watch('dir_serial', function(newVal, oldVal) {
				this.doSearch();
			});
			this.unWatchCurrentPage = this.$watch('currentPage', function(newVal, oldVal) {
				this.doSearch(newVal);
			});
			// this.getChannels();
			this.timer = setInterval(() => {
					this.getChannels(false);
			}, 3000);
			$(document).on("keydown", this.keyDown);
		},
		undoReady() {
			if(this.unWatchQ) {
				this.unWatchQ();
				this.unWatchQ = null;
			}
			if(this.unWatchOnline) {
				this.unWatchOnline();
				this.unWatchOnline = null;
			}
			if(this.unWatchChannelType) {
				this.unWatchChannelType();
				this.unWatchChannelType = null;
			}
			if(this.unWatchDirSerial) {
				this.unWatchDirSerial();
				this.unWatchDirSerial = null;
			}
			if(this.unWatchCurrentPage) {
				this.unWatchCurrentPage();
				this.unWatchCurrentPage = null;
			}
			if(this.timer) {
				clearInterval(this.timer);
				this.timer = 0;
			}
			$(document).off("keydown", this.keyDown);
		},
		keyDown(e) {
			if(e.keyCode == 27) {
				// this.$el.querySelector('.fa-chevron-left').click();
				if(this.xhr && typeof this.xhr.abort === 'function' && this.xhr.readyState !== 4) {
					this.xhr.abort();
				}
			}
		},
		doSearch(page = 1) {
			var query = {};
			if (this.q) query["q"] = this.q;
			if (this.type) query["type"] = this.type;
			if (this.online) query["online"] = this.online;
			if (this.dir_serial) query["dir_serial"] = this.dir_serial;
			this.$router.replace({
				path: `/devices/channels/${this.devid}/${page}`,
				query: query
			});
		},
		doDelaySearch: _.debounce(function() {
			this.doSearch();
		}, 800),
		getChannels(global = true) {
			if(global) {
				this.loading = true;
			} else {
				if(this.bgLoading || this.loading) return;
				this.bgLoading = true;
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/device/channellist",
				global: global,
				data: {
					serial: this.devid,
					q: this.q,
					start: (this.currentPage -1) * this.pageSize,
					limit: this.pageSize,
					online: this.online,
					channel_type: this.type,
					dir_serial: this.dir_serial,
					sort: this.sort,
					order: this.order
				}
			}).then(ret => {
				this.total = ret.ChannelCount;
				this.channels = ret.ChannelList;
			}).always(() => {
				if(global) {
					this.loading = false;
				} else {
					this.bgLoading = false;
				}
			});
		},
		uploaded() {
            this.$message({
                type: 'success',
                message: "上传成功！"
            })
            this.getChannels();
            this.$refs['uploadDlg'].hide();
		},
		sortChange(data) {
			this.sort = data.prop;
			this.order = data.order == "ascending" ? "asc" : "desc";
			this.getChannels();
		},
		canPlay(row) {
			// return row && row.DeviceOnline && row.Status == "ON" && !row.Custom && row.Parental == 0 && row.SubCount == 0;
			if (row && row.StreamID) return true;
			if (row && row.DeviceOnline && row.DeviceType == "GB" && (row.CustomStatus || row.Status) == "ON" && !row.Custom && row.SubCount == 0) {
				if (this.serverInfo.StrictChannelParental && row.Parental == 1) return false;
				if (this.serverInfo.StrictChannelCode && row.ID.length != 20) return false;
				if (row.Parental != 1 || row.Manufacturer != "LiveQing") {
					// if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215" && row.ID.substring(10, 13) != "200")) {
					if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215")) {
						return true;
					}
				}
			}
			return false;
		},
		canPlayback(row) {
			// return row && row.DeviceOnline && !row.Custom && row.Parental == 0 && row.SubCount == 0;
			if (row && row.DeviceOnline && row.DeviceType == "GB" && !row.Custom && row.SubCount == 0) {
				if (this.serverInfo.StrictChannelParental && row.Parental == 1) return false;
				if (this.serverInfo.StrictChannelCode && row.ID.length != 20) return false;
				if (row.Parental != 1 || row.Manufacturer != "LiveQing") {
					// if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215" && row.ID.substring(10, 13) != "200")) {
					if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215")) {
						return true;
					}
				}
			}
			return false;
		},
		canDecode(row) {
			// return row && row.DeviceOnline && row.Status == "ON" && !row.Custom && row.Parental == 0 && row.SubCount == 0;
			if (row && row.DeviceOnline && row.DeviceType == "Decode" && (row.CustomStatus || row.Status) == "ON" && !row.Custom && row.SubCount == 0) {
				if (this.serverInfo.StrictChannelParental && row.Parental == 1) return false;
				if (this.serverInfo.StrictChannelCode && row.ID.length != 20) return false;
				if (row.Parental != 1 || row.Manufacturer != "LiveQing") {
					// if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215" && row.ID.substring(10, 13) != "200")) {
					if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215")) {
						return true;
					}
				}
			}
			return false;
		},
		canCloudRecord(row) {
			// return row && row.CloudRecord && !row.Custom && row.Parental == 0 && row.SubCount == 0;
			if (row && row.CloudRecord && row.DeviceType == "GB" && !row.Custom && row.SubCount == 0) {
				if (this.serverInfo.StrictChannelParental && row.Parental == 1) return false;
				if (this.serverInfo.StrictChannelCode && row.ID.length != 20) return false;
				if (row.Parental != 1 || row.Manufacturer != "LiveQing") {
					// if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215" && row.ID.substring(10, 13) != "200")) {
					if (row.ID.length != 20 || (row.ID.substring(10, 13) != "216" && row.ID.substring(10, 13) != "215")) {
						return true;
					}
				}
			}
			return false;
		},
		canShare(row) {
			return (row && row.Shared && !this.serverInfo.GlobalChannelShared) || this.serverInfo.APIAuth === false || this.serverInfo.GlobalChannelShared === 1 || this.isDemoUser(this.serverInfo, this.userInfo);
		},
		isDir(row) {
			// return row && (row.SubCount > 0 || row.Parental == 1);
			if (row) {
				if (row.SubCount > 0) return true;
				if (this.serverInfo.StrictChannelParental && row.Parental == 1) return true;
				if (this.serverInfo.StrictChannelCode && row.ID.length <= 10) return true;
				if (row.Parental == 1 && row.Manufacturer == "LiveQing") return true;
				// if (row.ID.length == 20 && (row.ID.substring(10, 13) == "216" || row.ID.substring(10, 13) == "215" || row.ID.substring(10, 13) == "200")) {
				if (row.ID.length == 20 && (row.ID.substring(10, 13) == "216" || row.ID.substring(10, 13) == "215")) {
					return true;
				}
			}
			return false;
		},
		canFetchCatalog(row) {
			if(this.isDir(row) && row.DeviceOnline && row.ID != this.devid) {
				return true;
			}
			return false;
		},
		formatName(row, col, cell) {
			return row.CustomName || row.Name || "-";
		},
		formatPTZType(row, col, cell) {
			var ret = "未知";
			var val = row.CustomPTZType || row.PTZType;
			switch(val) {
				case 1:
					ret = "球机"; break;
				case 2:
					ret = "半球"; break;
				case 3:
					ret = "固定枪机"; break;
				case 4:
					ret = "遥控枪机"; break;
			}
			return ret;
		},
		formatPosition(row, col, cell) {
			var ret = "-,-";
			if(row.CustomLongitude || row.CustomLatitude) {
				ret = `${row.CustomLongitude},${row.CustomLatitude}`;
			} else if (row.Longitude || row.Latitude) {
				ret = `${row.Longitude},${row.Latitude}`;
			}
			return ret;
		},
		formatCatalogTitle(row) {
			if(this.catalogMap[`${this.devid}:${row.ID}`]) return "更新中...";
			return "";
		},
		setPTZType(cmd) {
			var ptz = cmd.PTZType;
			if(ptz === cmd.row.PTZType) {
				ptz = 0;
			}
			cmd.row.CustomPTZType = ptz;
			$.post("/api/v1/device/setchannelptztype", {
				serial: cmd.row.DeviceID,
				code: cmd.row.ID,
				ptz_type: ptz,
			})
		},
		toggleAudio(row) {
			$.post("/api/v1/device/setchannelaudio", {
				serial: row.DeviceID,
				code: row.ID,
				audio: !row.AudioEnable,
			}).then(() => {
				row.AudioEnable = !row.AudioEnable;
			})
		},
		editChannelName(row) {
			this.$refs['channelNameEditDlg'].show({
				serial: row.DeviceID,
				code: row.ID,
				name: row.Name,
				custom: row.Custom,
				customID: row.CustomID,
				customName: row.CustomName,
				description: row.Description,
			});
		},
		editChannelPosition(row) {
			this.$refs['channelPositionEditDlg'].show({
				serial: row.DeviceID,
				code: row.ID,
				name: row.CustomName || row.Name,
				custom: row.Custom,
				longitude: row.Longitude,
				latitude: row.Latitude,
				customLongitude: row.CustomLongitude,
				customLatitude: row.CustomLatitude,
			});
		},
		toggleOndemand(row) {
			$.post("/api/v1/device/setchannelondemand", {
				serial: row.DeviceID,
				code: row.ID,
				ondemand: !row.Ondemand,
			}).then(() => {
				row.Ondemand = !row.Ondemand;
			})
		},
		toggleCloudRecord(row) {
			var tip = row.CloudRecord ? "关闭云端录像, 是否同时设置 开启 按需直播?" : "开启云端录像, 是否同时设置 关闭 按需直播?";
			this.$confirm(tip, "提示", {
				distinguishCancelAndClose: true,
				cancelButtonText: "不设置",
				lockScroll: false,
			}).then(() => {
				 $.post("/api/v1/device/setchannelcloudrecord", {
					serial: row.DeviceID,
					code: row.ID,
					cloudrecord: !row.CloudRecord,
					ondemand: row.CloudRecord,
				}).then(() => {
					row.Ondemand = row.CloudRecord;
					row.CloudRecord = !row.CloudRecord;
				});
			}).catch(action => {
				if(action === 'cancel') {
					$.post("/api/v1/device/setchannelcloudrecord", {
						serial: row.DeviceID,
						code: row.ID,
						cloudrecord: !row.CloudRecord,
					}).then(() => {
						row.CloudRecord = !row.CloudRecord;
					})
				}
			});
		},
		toggleShared(row) {
			// 开启分享时，第一次给出安全提示
			if (!row.Shared && !this.sharedSecurityTipShown) {
				this.$confirm("开启分享后，该通道视频可在公开未登录的情况下被他人访问，请注意数据安全！确定开启吗？", "安全提示", {
					type: "warning",
					distinguishCancelAndClose: true,
					confirmButtonText: "确定开启",
					cancelButtonText: "取消",
					lockScroll: false,
				}).then(() => {
					this.sharedSecurityTipShown = true;
					this.doToggleShared(row);
				}).catch(() => {});
				return;
			}
			this.doToggleShared(row);
		},
		doToggleShared(row) {
			$.post("/api/v1/device/setchannelshared", {
				serial: row.DeviceID,
				code: row.ID,
				shared: !row.Shared,
			}).then(() => {
				row.Shared = !row.Shared;
			})
		},
		playStream(row) {
			this.loading = true;
			this.$set(row, "Locked", true);
			var dlg = this.$refs["videoDlg"];
			if (dlg) {
				dlg.videoTitle = row.CustomName || row.Name || row.ID;
				dlg.snapUrl = row.SnapURL || "";
				dlg.videoUrl = "";
				dlg.protocol = "";
				$(dlg.$el).modal("show");
			}
			this.xhr = $.post("/api/v1/stream/start", {
				serial: row.DeviceID,
				code: row.ID
			});
			this.xhr.then(streamInfo => {
				this.$refs["videoDlg"].play(row.CustomName||row.Name||row.ID, row.DeviceID, row.ID, streamInfo);
			}).fail(() => {
				if (dlg) $(dlg.$el).modal("hide");
			}).always(() => {
				this.xhr = null;
				this.loading = false;
				this.$delete(row, "Locked");
			})
		},
		decodeStart(row) {
			var strs = (row.StreamID||"").split(":");
			if(strs.length == 3) {
				this.loading = true;
				this.$set(row, "Locked", true);
				$.post("/api/v1/stream/start", {
					serial: strs[1],
					code: strs[2]
				}).then(streamInfo => {
					this.$refs["videoDlg"].play(row.CustomName||row.Name||row.ID, strs[1], strs[2], streamInfo);
				}).always(() => {
					this.loading = false;
					this.$delete(row, "Locked");
				})
			} else {
				this.$refs['decodeChannelListDlg'].show(row);
			}
		},
		decodeStream(dstRow, srcRow) {
			this.loading = true;
			this.$set(dstRow, "Locked", true);
			$.post("/api/v1/stream/start", {
				serial: srcRow.DeviceID,
				code: srcRow.ID
			}).then(streamInfo => {
				$.post("/api/v1/decode/start", {
					serial: dstRow.DeviceID,
					code: dstRow.ID,
					streamid: streamInfo.StreamID
				}).then(() => {
					this.$refs["videoDlg"].play(dstRow.CustomName||dstRow.Name||dstRow.ID, srcRow.DeviceID, srcRow.ID, streamInfo);
				}).always(() => {
					this.loading = false;
					this.$delete(dstRow, "Locked");
				})
			}).fail(() => {
				this.loading = false;
				this.$delete(dstRow, "Locked");
			})
		},
		stopStream(row) {
			if(row.DeviceType == "Decode") {
				this.$confirm(`确认停止解码 ${row.Name || row.ID} ?`, "提示", {
					lockScroll: false,
				}).then(() => {
					$.post("/api/v1/decode/stop", {
						serial: this.devid,
						code: row.ID
					}).always(() => {
						this.getChannels();
					});
				}).catch(() => {});
				return;
			}
			this.$confirm(`确认停止直播 ${row.Name || row.ID} ?`, "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/stream/stop", {
					serial: this.devid,
					code: row.ID
				}).always(() => {
					this.getChannels();
				});
			}).catch(() => {});
		},
		fetchCatalog(row) {
			if(this.serverInfo && !this.serverInfo.RemainDays) {
				this.$message({
					type: "error",
					message: "授权过期"
				})
				return
			}
			this.$set(this.catalogMap, `${this.devid}:${row.ID}`, true);
			$.ajax({
				method: "GET", // type: "GET", is also ok, method since jquery 1.9
				url: "/api/v1/device/fetchcatalog",
				global: false,
				data: {
					serial: this.devid,
					code: row.ID
				}
			}).then(() => {
				this.$message({
					type: "success",
					message: "更新通道信息成功"
				})
			}).fail(xhr => {
				xhr && console.log(`fetch catalog ajax error: ${xhr.status} ${xhr.responseText}`);
			}).always(() => {
				this.$delete(this.catalogMap, `${this.devid}:${row.ID}`);
			})
		},
		removeVirtual(row) {
			this.$confirm(`确认删除虚拟通道 ${row.Name || row.ID} ?`, "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/channel/virtual/remove", {
					serial: this.devid,
					code: row.ID
				}).always(() => {
					this.getChannels();
				})
			}).catch(() => {});
		},
		showDeviceLog(row) {
			this.$refs["deviceLogDlg"].show(`通道(${row.ID})`, row.DeviceID, row.ID);
		},
		showStatusLog(row) {
			this.$refs["deviceLogDlg"].showStatusLog(`通道(${row.ID})`, row.DeviceID, row.ID);
		},
		showStreamLog(row) {
			this.$refs["deviceLogDlg"].showStreamLog(`通道(${row.ID})`, row.DeviceID, row.ID);
		},
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.q = to.query.q || "";
			vm.online = to.query.online || "";
			vm.type = to.query.type || "";
			vm.dir_serial = to.query.dir_serial || "";
			vm.currentPage = parseInt(to.params.page) || 1;
			vm.ready();
		});
	},
	beforeRouteLeave(to, from, next) {
		this.undoReady();
		next();
	},
	beforeRouteUpdate(to, from, next) {
		this.undoReady();
		next();
		this.$nextTick(() => {
			this.q = to.query.q || "";
			this.online = to.query.online || "";
			this.type = to.query.type || "";
			this.dir_serial = to.query.dir_serial || "";
			this.currentPage = parseInt(to.params.page) || 1;
			this.channels = [];
			this.total = 0;
			this.getChannels();
			this.ready();
		});
	}
};
</script>

<style lang="less">
.opt-group .cell {
	overflow: visible;
}
</style>
