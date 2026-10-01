<template>
    <div>
        <div class="box box-primary">
            <div class="box-header">
                <h4 class="text-primary text-center">日志列表</h4>
            </div>
            <div class="box-body">
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group form-group-sm">
                        <label>搜索</label>
						<div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
							<input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
							<div class="input-group-btn">
								<button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['logExportDlg'].show()" title="导出日志列表"><i class="fa fa-download"></i></button>
							</div>
						</div>
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>调用方式</label>
                        <select class="form-control" v-model.trim="method">
                            <option value="">全部</option>
                            <option value="GET">HTTP GET</option>
                            <option value="POST">HTTP POST</option>
                            <option value="MESSAGE">SIP MESSAGE</option>
                            <option value="SUBSCRIBE">SIP SUBSCRIBE</option>
                            <option value="INFO">SIP INFO</option>
                            <option value="INVITE">SIP INVITE</option>
                            <option value="BYE">SIP BYE</option>
                        </select>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group">
                        <div class="input-group input-group-sm">
                            <DatePicker class="form-control input-group-sm" clearBtn :day.sync="starttime" ref="startDatePicker" placeholder="开始日期"></DatePicker>
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-default" @click.prevent="showStartDatePicker">
                                    <i class="fa fa-calendar"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                    <span class="hidden-xs">&nbsp;-&nbsp;</span>
                    <div class="form-group">
                        <div class="input-group input-group-sm">
                            <DatePicker class="form-control input-group-sm" clearBtn :day.sync="endtime" ref="endDatePicker" placeholder="结束日期"></DatePicker>
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-default" @click.prevent="showEndDatePicker">
                                    <i class="fa fa-calendar"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                    <div class="form-group form-group-sm pull-right" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员') && hasAllChannel(serverInfo, userInfo)">
                        <div class="input-group input-group-sm">
                            <button type="button" class="btn btn-sm btn-danger" @click.prevent="clearLog">
                                <i class="fa fa-remove"></i> 清空
                            </button>
                        </div>
                    </div>
                </div>
                <br>
                <div class="clearfix"></div>
                <el-table :data="logs" stripe :default-sort="{prop: 'StartAt', order: 'descending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column prop="Name" label="操作名称" min-width="140" show-overflow-tooltip :fixed="!isMobile()"></el-table-column>
                    <el-table-column prop="Method" label="调用方式" min-width="140" :formatter="formatMethod"></el-table-column>
                    <el-table-column prop="RequestURI" label="资源路径" min-width="200" show-overflow-tooltip>
						<template slot-scope="props">
							<a role="button" @click.prevent="alertExtInfo(props.row)" v-if="props.row.ExtInfo || props.row.Description">{{props.row.RequestURI}}</a>
							<span v-else>{{props.row.RequestURI}}</span>
						</template>
					</el-table-column>
                    <el-table-column prop="RemoteAddr" label="远端地址" min-width="150" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="RemoteRegion" label="所在地区" min-width="150" show-overflow-tooltip v-if="region"></el-table-column>
                    <el-table-column prop="Status" label="结果" min-width="140" :formatter="formatStatus" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Duration" label="用时(秒)" min-width="100" :formatter="formatDuration" sortable="custom" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Username" label="操作人" min-width="140" show-overflow-tooltip :fixed="isMobile() ? false : 'right'"></el-table-column>
                    <el-table-column prop="StartAt" label="操作时间" min-width="160" sortable="custom" :fixed="isMobile() ? false : 'right'"></el-table-column>
                </el-table>
            </div>
            <div class="box-footer" v-if="total > 0">
                <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
            </div>
        </div>
		<LogExportDlg ref="logExportDlg" @submit="exportLogs"/>
    </div>
</template>

<script>
import _ from "lodash";
import moment from "moment";
import DatePicker from "components/DatePicker.vue";
import LogExportDlg from "components/LogExportDlg.vue";
import { mapState } from "vuex";

const watchers = [];
export default {
	data() {
		return {
			q: "",
			method: "",
			starttime: "",
			endtime: "",
			total: 0,
			pageSize: 10,
			currentPage: 1,
			sort: "StartAt",
			order: "desc",
			loading: false,
			region: false,
			timer: 0,
			logs: [],
		};
	},
	computed: {
		...mapState(['userInfo', 'serverInfo']),
	},
	components: { DatePicker, LogExportDlg },
	mounted() {
		// this.timer = setInterval(() => {
		//   this.getLogs();
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
			watchers.push(this.$watch('q', function(newVal, oldVal) {
				this.doDelaySearch();
			}));
			watchers.push(this.$watch('starttime', function(newVal, oldVal) {
				this.doSearch();
			}));
			watchers.push(this.$watch('endtime', function(newVal, oldVal) {
				this.doSearch();
			}));
			watchers.push(this.$watch('method', function(newVal, oldVal) {
				this.doSearch();
			}));
			watchers.push(this.$watch('currentPage', function(newVal, oldVal) {
				this.doSearch(newVal);
			}));
			// this.getLogs(); // load when sort change
		},
		doSearch(page = 1) {
			var query = {};
			if (this.q) query["q"] = this.q;
			if (this.starttime) query["starttime"] = this.starttime;
			if (this.endtime) query["endtime"] = this.endtime;
			if (this.method) query["method"] = this.method;
			this.$router.replace({
				path: `/logs/${page}`,
				query: query
			});
		},
		doDelaySearch: _.debounce(function() {
			this.doSearch();
		}, 800),
		getLogs(global = true) {
			if(global) {
				this.loading = true;
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/log/list",
				global: global,
				data: {
					q: this.q,
					start: (this.currentPage -1) * this.pageSize,
					limit: this.pageSize,
					starttime: this.starttime ? moment(this.starttime, "YYYYMMDD").startOf('day').format("YYYYMMDDHHmmss") : "",
					endtime: this.endtime ? moment(this.endtime, "YYYYMMDD").endOf('day').format("YYYYMMDDHHmmss") : "",
					method: this.method,
					sort: this.sort,
					order: this.order
				}
			}).then(ret => {
				this.total = ret.LogCount;
				this.logs = ret.LogList;
				this.region = !!ret.LogRegion;
			}).always(() => {
				if(global) {
					this.$nextTick(() => {
						this.loading = false;
					})
				}
			});
		},
		exportLogs(filter = true, start = 0, limit = 10000) {
			let params = {
				start: start,
				limit: limit,
				sort: this.sort,
				order: this.order
			}
			if(filter) {
				params = Object.assign(params, {
					q: this.q,
					starttime: this.starttime ? moment(this.starttime, "YYYYMMDD").startOf('day').format("YYYYMMDDHHmmss") : "",
					endtime: this.endtime ? moment(this.endtime, "YYYYMMDD").endOf('day').format("YYYYMMDDHHmmss") : "",
					method: this.method
				})
			}
			let queryString = Object.keys(params).map(key => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`).join('&');
			window.open(`/api/v1/log/export?${queryString}`, "_blank");
		},
		sortChange(data) {
			this.sort = data.prop;
			this.order = data.order == "ascending" ? "asc" : "desc";
			this.getLogs();
		},
		formatMethod(row, col, cell) {
			return `${row.Scheme} ${row.Method}`;
		},
		formatStatus(row, col, cell) {
			return `${row.StatusCode} ${row.Status}`;
		},
		formatDuration(row, col, cell) {
			return moment.duration(row.Duration).asSeconds();
		},
		showStartDatePicker() {
			$(this.$refs.startDatePicker.$el).focus();
		},
		showEndDatePicker() {
			$(this.$refs.endDatePicker.$el).focus();
		},
		removeLog(row) {
			this.$confirm(`确认删除`, "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/log/remove", {
					id: row.ID
				}).always(() => {
					this.getLogs();
				});
			}).catch(() => {});
		},
		clearLog() {
			this.$confirm('确认全部删除', "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/log/clear", {}).then(() => {
					this.getLogs();
				})
			}).catch(() => {});
		},
		alertExtInfo(row) {
			let info = "";
			if(row.Scheme == "SIP" && row.ExtInfo) {
				if(row.ExtInfo != this.q && ("Call-ID:"+row.ExtInfo) != this.q) {
					info = `Call-ID: <a href='#/logs/1?q=${encodeURIComponent("Call-ID:")}${encodeURIComponent(row.ExtInfo)}'>${this.escapeHtml(row.ExtInfo)}</a>`;
				} else {
					info = this.escapeHtml(`Call-ID: ${row.ExtInfo}`);
				}
			} else if(row.Scheme == "HTTP" && row.Method == "POST") {
				info = `POST ${this.escapeHtml(row.RequestURI)}`;
			}
			if(row.Description) {
				if(row.Scheme == "SIP" && row.Method == "INVITE") {
					let match = row.Description.match(/^t=(\d+)\s+(\d+)/m);
					if(match && match.length == 3) {
						let t1 = parseInt(match[1]) || 0;
						let t2 = parseInt(match[2]) || 0;
						if(t1 && t2) {
							if(info) {
								info += "\r\n";
							}
							info += `Time-Range: ${moment.unix(t1).format('YYYY-MM-DD HH:mm:ss')} ~ ${moment.unix(t2).format('YYYY-MM-DD HH:mm:ss')}`;
						}
					}
				}
				if(info) {
					info += "\r\n\r\n";
				}
				info += `${this.escapeHtml(row.Description)}`;
			}
			this.$alert(info.replace(/\r?\n/g, "<br>"), row.Name||"提示", {
				lockScroll: false,
				showConfirmButton: false,
				dangerouslyUseHTMLString: true,
				closeOnClickModal: false,
				closeOnPressEscape: true,
				customClass: "break-word",
			}).catch(() => {});
		},
		escapeHtml(str) {
			return str.replace(/&/g, '&amp;')
				.replace(/ /g, '&nbsp;')
				.replace(/\t/g, '&nbsp;&nbsp;')
				.replace(/</g, '&lt;')
				.replace(/>/g, '&gt;')
				.replace(/"/g, '&quot;')
				.replace(/'/g, '&#039;');
		}
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.q = to.query.q || "";
			vm.starttime = to.query.starttime || "";
			vm.endtime = to.query.endtime || "";
			vm.method = to.query.method || "";
			vm.total = 0;
			vm.currentPage = parseInt(to.params.page) || 1;
			vm.ready();
		});
	},
	beforeRouteUpdate(to, from, next) {
        watchers.forEach(unwatch => unwatch());
        watchers.length = 0;
		next();
		this.q = to.query.q || "";
		this.starttime = to.query.starttime || "";
		this.endtime = to.query.endtime || "";
		this.method = to.query.method || "";
		this.total = 0;
		this.currentPage = parseInt(to.params.page) || 1;
		this.getLogs();
		this.ready();
	}
};
</script>
