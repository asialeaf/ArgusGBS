<template>
    <div>
        <div class="box box-primary">
            <div class="box-header">
                <h4 class="text-primary text-center">上级平台级联列表</h4>
            </div>
            <div class="box-body">
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group form-group-sm">
                        <div class="input-group input-group-sm">
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['cascadeEditDlg'].show()">
                                <i class="fa fa-plus"></i> 添加上级平台
                            </button>
                        </div>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>搜索</label>
                        <div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                            <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出级联列表"><i class="fa fa-download"></i></button>
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['uploadDlg'].show()" title="导入级联列表"><i class="fa fa-upload"></i></button>
                            </div>
                        </div>
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>开启状态</label>
                        <select class="form-control" v-model.trim="enable">
                            <option value="">全部</option>
                            <option value="true">启用</option>
                            <option value="false">关闭</option>
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
                <div class="clearfix"></div>
                <el-table :data="cascades" stripe :default-sort="{prop: 'CreatedAt', order: 'descending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column prop="Name" label="名称" min-width="140" show-overflow-tooltip></el-table-column>
                    <el-table-column label="操作" min-width="300" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                        <template slot-scope="props">
                            <div class="btn-group btn-group-xs">
                                <button type="button" class="btn btn-warning" @click.prevent="editCascade(props.row)">
                                    <i class="fa fa-edit"></i> 编辑
                                </button>
                                <button type="button" class="btn btn-primary" @click.prevent="editChannel(props.row)">
                                    <i class="fa fa-check"></i> 选择通道
                                </button>
                                <button type="button" class="btn btn-info" @click.prevent="pushChannel(props.row)" v-if="props.row.Online">
                                    <i class="fa fa-angle-double-up"></i> 推送通道
                                </button>
                                <button type="button" class="btn btn-danger" @click.prevent="removeCascade(props.row)">
                                    <i class="fa fa-remove"></i> 删除
                                </button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column prop="Serial" label="SIP服务国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                    <!-- <el-table-column prop="Realm" label="SIP服务国标域" min-width="120" show-overflow-tooltip></el-table-column> -->
                    <el-table-column prop="Enable" label="是否启用" min-width="100">
                        <template slot-scope="props">
                            <el-switch :value="props.row.Enable" @change="toggleEnable(props.row)"></el-switch>
                        </template>
                    </el-table-column>
                    <el-table-column prop="Online" label="在线状态" min-width="100">
                        <template slot-scope="props">
                            <span v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')">
                                <a role="button" @click.prevent="showStatusLog(props.row)"
                                    :class="[{'text-success': props.row.Online, 'text-gray': !props.row.Online}]" title="点击查看状态记录">
                                    {{props.row.Online ? "在线" : "离线"}}
                                </a>
                            </span>
                            <span v-else-if="props.row.Online" class="text-success">在线</span>
                            <span v-else class="text-gray">离线</span>
							<el-popover trigger="hover" v-if="props.row.Online && (props.row.CatalogSubscribed || props.row.AlarmSubscribed || props.row.PositionSubscribed || props.row.PTZSubscribed)">
								<p :class="{'text-success': props.row.CatalogSubscribed}">目录订阅 <i :class="['pull-right', 'fa', {'fa-check': props.row.CatalogSubscribed, 'fa-close': !props.row.CatalogSubscribed}]"></i></p>
								<p :class="{'text-success': props.row.AlarmSubscribed}">报警订阅 <i :class="['pull-right', 'fa', {'fa-check': props.row.AlarmSubscribed, 'fa-close': !props.row.AlarmSubscribed}]"></i></p>
								<p :class="{'text-success': props.row.PositionSubscribed}">位置订阅 <i :class="['pull-right', 'fa', {'fa-check': props.row.PositionSubscribed, 'fa-close': !props.row.PositionSubscribed}]"></i></p>
								<p :class="{'text-success': props.row.PTZSubscribed}" style="margin-bottom:0;">PTZ订阅 <i :class="['pull-right', 'fa', {'fa-check': props.row.PTZSubscribed, 'fa-close': !props.row.PTZSubscribed}]"></i></p>
								<span slot="reference" class="text-success">&nbsp;&nbsp;<i class="fa fa-rss"></i></span>
							</el-popover>
                        </template>
                    </el-table-column>
                    <el-table-column prop="Status" label="状态描述" min-width="150" show-overflow-tooltip>
                        <template slot-scope="props">
                            <a role="button" class="text-success" @click="showSession(props.row)" title="点击查看会话列表"
                                v-if="props.row.Online && props.row.Load && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                                {{props.row.Status}}, Load[{{props.row.Load}}]
                            </a>
                            <span v-else>{{props.row.Status}}</span>
                        </template>
                    </el-table-column>
                    <el-table-column prop="Host" label="SIP服务IP" min-width="140" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Port" label="SIP服务端口" min-width="120"></el-table-column>
                    <el-table-column prop="LocalSerial" label="设备国标编号" min-width="200" show-overflow-tooltip></el-table-column>
                    <!-- <el-table-column prop="RegisterTimeout" label="注册有效期(秒)" min-width="120"></el-table-column> -->
                    <el-table-column prop="RegisterInterval" label="注册周期(秒)" min-width="120"></el-table-column>
                    <el-table-column prop="KeepaliveInterval" label="心跳周期(秒)" min-width="120"></el-table-column>
                    <el-table-column prop="CatalogGroupSize" label="目录分组大小" min-width="120"></el-table-column>
                    <el-table-column prop="CommandTransport" label="信令传输" min-width="120"></el-table-column>
                    <el-table-column prop="Charset" label="字符集" min-width="120"></el-table-column>
                    <el-table-column prop="UpdatedAt" label="更新时间" min-width="160" sortable="custom"></el-table-column>
                    <el-table-column prop="CreatedAt" label="创建时间" min-width="160" sortable="custom"></el-table-column>
                </el-table>
            </div>
            <div class="box-footer" v-if="total > 0">
                <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
            </div>
        </div>
        <UploadDlg ref="uploadDlg" title="上传级联列表" url="/api/v1/cascade/import" @uploaded="uploaded"></UploadDlg>
        <CascadeEditDlg ref="cascadeEditDlg" :serverInfo="serverInfo" :userInfo="userInfo" @submit="getCascades"></CascadeEditDlg>
        <CascadeChannelListDlg ref="cascadeChannelListDlg" :serverInfo="serverInfo" :userInfo="userInfo" :title="cascadeChannelListDlgTitle"></CascadeChannelListDlg>
        <CascadeSessionListDlg ref="cascadeSessionListDlg" :serverInfo="serverInfo" :userInfo="userInfo" :title="cascadeSessionListDlgTitle"></CascadeSessionListDlg>
        <CascadeLogDlg ref="cascadeLogDlg"></CascadeLogDlg>
    </div>
</template>

<script>
import _ from "lodash";
import UploadDlg from "components/UploadDlg.vue";
import CascadeEditDlg from "components/CascadeEditDlg.vue";
import CascadeChannelListDlg from "components/CascadeChannelListDlg.vue";
import CascadeSessionListDlg from "components/CascadeSessionListDlg.vue";
import CascadeLogDlg from "components/CascadeLogDlg.vue";
import { mapState } from "vuex";
export default {
	data() {
		return {
			q: "",
			online: "",
			enable: "",
			total: 0,
			pageSize: 10,
			currentPage: 1,
			sort: "Serial",
			order: "asc",
			loading: false,
			bgLoading: false,
			timer: 0,
			cascades: [],
			cascadeChannelListDlgTitle: "选择通道",
			cascadeSessionListDlgTitle: "级联会话列表",
		};
	},
	computed: {
		...mapState(['userInfo', 'serverInfo']),
	},
	components: {
		UploadDlg, CascadeEditDlg, CascadeChannelListDlg, CascadeSessionListDlg, CascadeLogDlg
	},
	mounted() {
		// this.getCascades();
	},
	beforeDestroy() {
		if (this.timer) {
			clearInterval(this.timer);
			this.timer = 0;
		}
	},
	methods: {
		ready() {
            this.$watch('q', function(newVal, oldVal) {
                this.doDelaySearch();
            });
            this.$watch('online', function(newVal, oldVal) {
                this.doSearch();
            });
            this.$watch('enable', function(newVal, oldVal) {
                this.doSearch();
            })
            this.$watch('currentPage', function(newVal, oldVal) {
                this.doSearch(newVal);
            });
            // this.getCascades();
            this.timer = setInterval(() => {
                this.getCascades(false);
            }, 3000);
		},
		doSearch(page = 1) {
			var query = {};
			if (this.q) query["q"] = this.q;
			if (this.online) query["online"] = this.online;
			if (this.enable) query["enable"] = this.enable;
			this.$router.replace({
				path: `/cascade/${page}`,
				query: query
			});
		},
		doDelaySearch: _.debounce(function() {
			this.doSearch();
		}, 800),
		getCascades(global = true) {
			if(global) {
				this.loading = true;
			} else {
				if(this.bgLoading || this.loading) return;
				this.bgLoading = true;
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/cascade/list",
				global: global,
				data: {
					q: this.q,
					start: (this.currentPage -1) * this.pageSize,
					limit: this.pageSize,
					online: this.online,
					enable: this.enable,
					sort: this.sort,
					order: this.order
				}
			}).then(ret => {
				this.total = ret.CascadeCount;
				this.cascades = ret.CascadeList;
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
			this.getCascades();
		},
		formatStatus(row, col, cell) {
			var ret = cell;
			if(row.Online && row.Load) {
				if(ret) {
					ret += `, Load[${row.Load}]`;
				} else {
					ret = `Load[${row.Load}]`;
				}
			}
			// if(row.Manufacturer) {
			//   if(ret) {
			//     ret += `, ${row.Manufacturer}`;
			//   } else {
			//     ret = `${row.Manufacturer}`;
			//   }
			// }
			return ret;
		},
		editCascade(row) {
			this.$refs['cascadeEditDlg'].show(row);
		},
		editChannel(row) {
			this.cascadeChannelListDlgTitle = `选择通道(${row.Name || row.ID})`;
			this.$refs['cascadeChannelListDlg'].show(row.ID, row.Name);
		},
		showSession(row) {
			this.cascadeSessionListDlgTitle = `级联会话列表(${row.Name || row.ID})`;
			this.$refs['cascadeSessionListDlg'].show(row.ID, row.Serial, row.Name);
		},
		showStatusLog(row) {
			this.$refs["cascadeLogDlg"].show(row);
		},
		pushChannel(row) {
			if(!row.Online) {
				this.$message({
					type: "error",
					message: "平台不在线!"
				})
				return
			}
			$.post("/api/v1/cascade/pushcatalog", {
				id: row.ID
			}).then(() => {
				this.$message({
					type: "success",
					message: "推送通道信息成功"
				})
			})
		},
		removeCascade(row) {
			this.$confirm(`确认删除 ${row.Name || row.ID}`, "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/cascade/remove", {
					id: row.ID
				}).always(() => {
					this.getCascades();
				});
			}).catch(() => {});
		},
		download() {
			window.open(`/api/v1/cascade/export?q=${this.q}&enable=${this.enable}&online=${this.online}`);
		},
		uploaded() {
            this.$message({
                type: 'success',
                message: "上传成功！"
            })
            this.getCascades();
            this.$refs['uploadDlg'].hide();
		},
		toggleEnable(row) {
			$.post("/api/v1/cascade/setenable", {
				id: row.ID,
				enable: !row.Enable,
			}).then(() => {
				row.Enable = !row.Enable;
			})
		},
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.q = to.query.q || "";
			vm.online = to.query.online || "";
			vm.enable = to.query.enable || "";
			vm.currentPage = parseInt(to.params.page) || 1;
			vm.ready();
		});
	},
	beforeRouteUpdate(to, from, next) {
		next();
		this.$nextTick(() => {
			this.q = to.query.q || "";
			this.online = to.query.online || "";
			this.enable = to.query.enable || "";
			this.currentPage = parseInt(to.params.page) || 1;
			this.cascades = [];
			this.total = 0;
			this.getCascades();
		});
	}
};
</script>

<style lang="less">
.opt-group .cell {
	overflow: visible;
}
</style>
