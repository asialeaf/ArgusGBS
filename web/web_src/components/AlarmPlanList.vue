<template>
    <div>
        <div class="box box-primary">
            <div class="box-header">
                <h4 class="text-primary text-center">报警预案列表</h4>
            </div>
            <div class="box-body">
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group">
                        <button type="button" class="btn btn-primary btn-sm" @click.prevent="$router.go(-1)" v-if="$getQueryString('menu') != 'no'">
                            <i class="fa fa-chevron-left"></i> 返回
                        </button>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <div class="input-group input-group-sm">
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['alarmPlanEditDlg'].show()">
                                <i class="fa fa-plus"></i> 添加报警预案
                            </button>
                        </div>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>搜索</label>
                        <div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                            <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出报警预案列表"><i class="fa fa-download"></i></button>
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['uploadDlg'].show()" title="导入报警预案列表"><i class="fa fa-upload"></i></button>
                            </div>
                        </div>
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                    </div>
                </div>
                <br>
                <div class="clearfix"></div>
                <el-table :data="alarmPlans" stripe :default-sort="{prop: 'CreatedAt', order: 'descending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column prop="Name" label="名称" min-width="150" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column label="操作" min-width="250" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                        <template slot-scope="props">
                            <div class="btn-group btn-group-xs">
                                <button type="button" class="btn btn-warning" @click.prevent="editAlarmPlan(props.row)">
                                    <i class="fa fa-edit"></i> 编辑
                                </button>
                                <button type="button" class="btn btn-primary" @click.prevent="editChannel(props.row)">
                                    <i class="fa fa-check"></i> 关联通道
                                </button>
                                <button type="button" class="btn btn-danger" @click.prevent="removeAlarmPlan(props.row)">
                                    <i class="fa fa-remove"></i> 删除
                                </button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column prop="Enable" label="是否启用" min-width="100" sortable="custom">
                        <template slot-scope="props">
                            <el-switch :value="props.row.Enable" @change="toggleEnable(props.row)"></el-switch>
                        </template>
                    </el-table-column>
                    <el-table-column prop="SnapInterval" label="快照周期(秒)" :formatter="formatSnapInterval"  min-width="160" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column prop="RecordDuration" label="录像时长(秒)" :formatter="formatRecordDurantion"  min-width="160"  show-overflow-tooltip sortable="custom" v-if="serverInfo.VersionType != '标准版'"></el-table-column>
                    <el-table-column prop="UpdatedAt" label="更新时间" min-width="160" sortable="custom"></el-table-column>
                    <el-table-column prop="CreatedAt" label="创建时间" min-width="160" sortable="custom"></el-table-column>
                </el-table>
            </div>
            <div class="box-footer" v-if="total > 0">
                <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
            </div>
        </div>
        <UploadDlg ref="uploadDlg" title="上传报警预案列表" url="/api/v1/alarm/plan/import" @uploaded="uploaded"></UploadDlg>
        <AlarmPlanEditDlg ref="alarmPlanEditDlg" :serverInfo="serverInfo" :userInfo="userInfo" @submit="getAlarmPlans()"></AlarmPlanEditDlg>
        <AlarmPlanChannelListDlg ref="alarmPlanChannelListDlg" :serverInfo="serverInfo" :userInfo="userInfo" :title="alarmPlanChannelListDlgTitle"></AlarmPlanChannelListDlg>
    </div>
</template>

<script>
import _ from "lodash";
import UploadDlg from "components/UploadDlg.vue";
import AlarmPlanEditDlg from "components/AlarmPlanEditDlg";
import AlarmPlanChannelListDlg from "components/AlarmPlanChannelListDlg";
import { mapState } from "vuex";
export default {
	data() {
		return {
			q: "",
			online: "",
			total: 0,
			pageSize: 10,
			currentPage: 1,
			sort: "Name",
			order: "asc",
			loading: false,
			bgLoading: false,
			timer: 0,
			alarmPlans: [],
			alarmPlanChannelListDlgTitle: "选择通道",
		};
	},
	computed: {
		...mapState(['userInfo', 'serverInfo']),
	},
	components: {
		UploadDlg, AlarmPlanEditDlg, AlarmPlanChannelListDlg
	},
	mounted() {
	},
	methods: {
		ready() {
            this.$watch('q', function(newVal, oldVal) {
                this.doDelaySearch();
            });
            this.$watch('online', function(newVal, oldVal) {
                this.doSearch(newVal);
            });
            this.$watch('currentPage', function(newVal, oldVal) {
                this.doSearch(newVal);
            });
            this.timer = setInterval(() => {
                this.getAlarmPlans(false);
            }, 3000);
		},
		doSearch(page = 1) {
			var query = {};
			if (this.q) query["q"] = this.q;
			if (this.online) query["online"] = this.online;
			this.$router.replace({
				path: `/alarms/plans/${page}`,
				query: query
			});
		},
		doDelaySearch: _.debounce(function() {
			this.doSearch();
		}, 800),
		getAlarmPlans(global = true) {
			if(global) {
				this.loading = true;
			} else {
				if(this.bgLoading || this.loading) return;
				this.bgLoading = true;
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/alarm/plan/list",
				global: global,
				data: {
					q: this.q,
					start: (this.currentPage -1) * this.pageSize,
					limit: this.pageSize,
					sort: this.sort,
					order: this.order
				}
			}).then(ret => {
				this.total = ret.PlanCount;
				this.alarmPlans = ret.PlanList;
			}).always(() => {
				if(global) {
					this.loading = false;
				} else {
					this.bgLoading = false;
				}
			});
		},
		formatRecordDurantion(row, col, cell) {
            if (cell == "0") {
                return "不录制视频";
            }
            return cell;
		},
		formatSnapInterval(row, col, cell) {
		    if (cell == "0") {
                return "不截取快照";
            }
            return cell;
		},
		sortChange(data) {
			this.sort = data.prop;
			this.order = data.order == "ascending" ? "asc" : "desc";
			this.getAlarmPlans();
		},
		formatName(row, col, cell) {
			if (cell) return cell;
			return "-";
		},
		editAlarmPlan(row) {
			this.$refs['alarmPlanEditDlg'].show(row);
		},
		editChannel(row) {
			this.alarmPlanChannelListDlgTitle = `报警预案关联通道(${row.Name || row.ID})`;
			this.$refs['alarmPlanChannelListDlg'].show(row.ID);
		},
		removeAlarmPlan(row) {
			this.$confirm(`确认删除 ${row.Name || row.ID}`, "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/alarm/plan/remove", {
					ID: row.ID
				}).always(() => {
					this.getAlarmPlans();
				});
			}).catch(() => {});
		},
		download() {
			window.open(`/api/v1/alarm/plan/export`);
		},
		uploaded() {
            this.$message({
                type: 'success',
                message: "上传成功！"
            })
            this.getAlarmPlans();
            this.$refs['uploadDlg'].hide();
		},
		toggleEnable(row) {
			$.post("/api/v1/alarm/plan/setenable", {
				ID: row.ID,
				Enable: !row.Enable,
			}).then(() => {
				row.Enable = !row.Enable;
			})
		},
	},
	beforeDestroy() {
		if (this.timer) {
			clearInterval(this.timer);
			this.timer = 0;
		}
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.q = to.query.q || "";
			vm.online = to.query.online || "";
			vm.currentPage = parseInt(to.params.page) || 1;
			vm.ready();
		});
	},
	beforeRouteUpdate(to, from, next) {
		next();
		this.$nextTick(() => {
			this.q = to.query.q || "";
			this.online = to.query.online || "";
			this.currentPage = parseInt(to.params.page) || 1;
			this.alarmPlans = [];
			this.total = 0;
			this.getAlarmPlans();
		});
	}
};
</script>

<style lang="less">
.opt-group .cell {
	overflow: visible;
}
</style>
