<template>
<div>
    <div class="box box-primary">
        <div class="box-header">
            <h4 class="text-primary text-center">录像计划列表</h4>
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
                        <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['cloudRecordEditDlg'].show()">
                            <i class="fa fa-plus"></i> 添加录像计划
                        </button>
                    </div>
                </div>
                <span class="hidden-xs">&nbsp;&nbsp;</span>
                <div class="form-group form-group-sm">
                    <label>搜索</label>
                    <div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                        <div class="input-group-btn">
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出录像计划列表"><i class="fa fa-download"></i></button>
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['uploadDlg'].show()" title="导入录像计划列表"><i class="fa fa-upload"></i></button>
                        </div>
                    </div>
                    <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                </div>
            </div>
            <br>
            <div class="clearfix"></div>
            <el-table :data="cloudRecords" stripe :default-sort="{prop: 'CreatedAt', order: 'descending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                <el-table-column prop="Name" label="名称" min-width="150" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column label="操作" min-width="220" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                    <template slot-scope="props">
                        <div class="btn-group btn-group-xs">
                            <button type="button" class="btn btn-warning" @click.prevent="editCloudRecordPlan(props.row)">
                                <i class="fa fa-edit"></i> 编辑
                            </button>
                            <button type="button" class="btn btn-primary" @click.prevent="editChannel(props.row)">
                                <i class="fa fa-check"></i> 关联通道
                            </button>
                            <button type="button" class="btn btn-danger" @click.prevent="removeCloudRecordPlan(props.row)">
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
                <el-table-column prop="Plan" label="录像计划详情" min-width="470" :formatter="formatPlan" show-overflow-tooltip></el-table-column>
                <el-table-column prop="UpdatedAt" label="更新时间" min-width="160" sortable="custom"></el-table-column>
                <el-table-column prop="CreatedAt" label="创建时间" min-width="160" sortable="custom"></el-table-column>
            </el-table>
        </div>
        <div class="box-footer" v-if="total > 0">
            <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
        </div>
    </div>
    <UploadDlg ref="uploadDlg" title="上传录像计划列表" url="/api/v1/cloudrecord/plan/import" @uploaded="uploaded"></UploadDlg>
    <CloudRecordPlanEditDlg ref="cloudRecordEditDlg" @submit="getCloudRecordPlans()"></CloudRecordPlanEditDlg>
    <CloudRecordPlanChannelListDlg ref="cloudRecordChannelListDlg" :serverInfo="serverInfo" :userInfo="userInfo" :title="cloudRecordChannelListDlgTitle"></CloudRecordPlanChannelListDlg>
</div>
</template>

<script>
import _ from "lodash";
import UploadDlg from "components/UploadDlg.vue";
import CloudRecordPlanEditDlg from "components/CloudRecordPlanEditDlg.vue";
import CloudRecordPlanChannelListDlg from "components/CloudRecordPlanChannelListDlg.vue";
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
            timer: 0,
            cloudRecords: [],
            cloudRecordChannelListDlgTitle: "选择通道",
        };
    },
    computed: {
        ...mapState(['userInfo', 'serverInfo']),
    },
    components: {
        UploadDlg, CloudRecordPlanEditDlg, CloudRecordPlanChannelListDlg
    },
    mounted() {
        // this.getCloudRecordPlans();
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
                this.doSearch(newVal);
            });
            this.$watch('currentPage', function(newVal, oldVal) {
                this.doSearch(newVal);
            });
            this.timer = setInterval(() => {
                this.getCloudRecordPlans(false);
            }, 3000);
        },
        doSearch(page = 1) {
            var query = {};
            if (this.q) query["q"] = this.q;
            if (this.online) query["online"] = this.online;
            this.$router.replace({
                path: `/cloudrecord/plans/${page}`,
                query: query
            });
        },
        doDelaySearch: _.debounce(function() {
            this.doSearch();
        }, 800),
        getCloudRecordPlans(global = true) {
            if(global) {
                this.loading = true;
            }
            $.ajax({
                method: "GET",
                url: "/api/v1/cloudrecord/plan/list",
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
                this.cloudRecords = ret.PlanList;
            }).always(() => {
                if(global) {
                    this.$nextTick(() => {
                        this.loading = false;
                    })
                }
            });
        },
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order == "ascending" ? "asc" : "desc";
            this.getCloudRecordPlans();
        },
        formatName(row, col, cell) {
            if (cell) return cell;
            return "-";
        },
        formatPlan(row, col, cell) {
            var c = "";
            var days = ['星期一','星期二','星期三','星期四','星期五','星期六','星期日'];
            for (var i = 0; i < 7; i++) {
                if (cell && cell.indexOf("s"+(i+1)) >= 0) {
                    c += days[i] + "[✓] ";
                } else {
                    c += days[i] + "[✗] ";
                }
            }
            return c;
        },
        editCloudRecordPlan(row) {
            this.$refs['cloudRecordEditDlg'].show(row);
        },
        editChannel(row) {
            this.cloudRecordChannelListDlgTitle = `录像计划关联通道(${row.Name || row.ID})`;
            this.$refs['cloudRecordChannelListDlg'].show(row.ID);
        },
        pushChannel(row) {
            if(!row.Online) {
                this.$message({
                    type: "error",
                    message: "平台不在线!"
                })
                return
            }
            $.get("/api/v1/cloudRecord/pushcatalog", {
                id: row.ID
            }).then(() => {
                this.$message({
                    type: "success",
                    message: "推送通道信息成功"
                })
            })
        },
        removeCloudRecordPlan(row) {
            this.$confirm(`确认删除 ${row.Name || row.ID}`, "提示", {
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/cloudrecord/plan/remove", {
                    ID: row.ID
                }).always(() => {
                    this.getCloudRecordPlans();
                });
            }).catch(() => {});
        },
        download() {
            window.open(`/api/v1/cloudrecord/plan/export?q=${this.q}`);
        },
        uploaded() {
            this.$message({
                type: 'success',
                message: "上传成功！"
            })
            this.getCloudRecordPlans();
            this.$refs['uploadDlg'].hide();
        },
        toggleEnable(row) {
            $.post("/api/v1/cloudrecord/plan/setenable", {
                ID: row.ID,
                Enable: !row.Enable,
            }).then(() => {
                row.Enable = !row.Enable;
            })
        },
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
            this.cloudRecords = [];
            this.total = 0;
            this.getCloudRecordPlans();
        });
    }
};
</script>

<style lang="less">
.opt-group .cell {
    overflow: visible;
}
</style>
