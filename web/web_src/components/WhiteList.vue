<template>
<div>
    <div class="box box-primary">
        <div class="box-header">
            <h4 class="text-primary text-center">白名单</h4>
        </div>
        <div class="box-body">
            <div class="form-inline" autocomplete="off" spellcheck="false">
                <div class="form-group form-group-sm">
                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="$router.go(-1)">
                        <i class="fa fa-chevron-left"></i> 返回
                    </button>
                </div>
                <span class="hidden-xs">&nbsp;&nbsp;</span>
                <div class="form-group form-group-sm">
                    <div class="input-group input-group-sm">
                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="addWhite()">
                        <i class="fa fa-plus"></i> 添加白名单
                    </button>
                    </div>
                </div>
                <span class="hidden-xs">&nbsp;&nbsp;</span>
                <div class="form-group form-group-sm">
                    <label>搜索</label>
                    <div class="input-group input-group-sm" v-if="!isMobile()">
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                        <div class="input-group-btn">
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出白名单数据"><i class="fa fa-download"></i></button>
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['uploadDlg'].show()" title="导入白名单数据"><i class="fa fa-upload"></i></button>
                        </div>
                    </div>
                    <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                </div>
            </div>
            <br>
            <el-table :data="whiteList" stripe :default-sort="{prop: 'Serial', order: 'ascending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                <el-table-column prop="Serial" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column label="操作" min-width="150" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                    <template slot-scope="props">
                        <div class="btn-group btn-group-xs">
                            <button type="button" class="btn btn-warning" @click.prevent="editWhite(props.row)" v-if="userInfo">
                                <i class="fa fa-edit"></i> 编辑
                            </button>
                            <button type="button" class="btn btn-danger" @click.prevent="delWhite(props.row)" v-if="userInfo">
                                <i class="fa fa-remove"></i> 删除
                            </button>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="Password" label="接入密码" min-width="140" :formatter="formatPassword" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column prop="UserAgent" label="UA" min-width="140" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column prop="RemoteIP" label="IP" min-width="140" :formatter="formatRemoteIP" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column prop="RemotePort" label="端口" min-width="140" :formatter="formatRemotePort" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column prop="Description" label="备注" min-width="200" show-overflow-tooltip sortable="custom">
                    <template slot-scope="props">
                        <span>{{props.row.Description}}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="UpdatedAt" label="更新时间" min-width="160" sortable="custom"></el-table-column>
                <el-table-column prop="CreatedAt" label="创建时间" min-width="160" sortable="custom"></el-table-column>
            </el-table>
        </div>
        <div class="box-footer" v-if="total > 0">
            <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
        </div>

        <UploadDlg ref="uploadDlg" title="上传白名单" url="/api/v1/white/import" @uploaded="uploaded"></UploadDlg>
        <WhiteEditDlg ref="whiteEditDlg" :serverInfo="serverInfo" :userInfo="userInfo" @submit="getWhiteList"></WhiteEditDlg>
    </div>
</div>
</template>

<script>
import UploadDlg from "components/UploadDlg.vue";
import WhiteEditDlg from "components/WhiteEditDlg.vue";
import _ from "lodash";
import { mapState } from "vuex";
export default {
    props: {},
    data() {
        return {
            q: "",
            total: 0,
            pageSize: 10,
            currentPage: 1,
            sort: "Serial",
            order: "asc",
            loading: false,
            whiteList: [],
        };
    },
    components: { UploadDlg, WhiteEditDlg },
    computed: {
        ...mapState(['userInfo', 'serverInfo']),
    },
    mounted() {
        // this.getWhiteList();
    },
    beforeDestroy() {
    },
    methods: {
        ready() {
            this.$watch('q', function(newVal, oldVal) {
                this.doDelaySearch();
            });
            this.$watch('currentPage', function(newVal, oldVal) {
                this.doSearch(newVal);
            });
            // this.getWhiteList();
        },
        download() {
            window.open(`/api/v1/white/export?q=${this.q}`);
        },
        uploaded() {
            this.$message({
                type: 'success',
                message: "上传配置成功！"
            })
            this.getWhiteList();
            this.$refs['uploadDlg'].hide();
        },
        doSearch(page = 1) {
            var query = {};
            if (this.q) query["q"] = this.q;
            this.$router.replace({
                path: `/white/${page}`,
                query: query
            });
        },
        doDelaySearch: _.debounce(function () {
            this.doSearch();
        }, 800),
        getWhiteList() {
            this.loading = true;
            $.get("/api/v1/white/list", {
                q: this.q,
                start: (this.currentPage - 1) * this.pageSize,
                limit: this.pageSize,
                sort: this.sort,
                order: this.order
            }).then(ret => {
                this.total = ret.WhiteCount;
                this.whiteList = ret.WhiteList;
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
            });
        },
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order == "ascending" ? "asc" : "desc";
            this.getWhiteList();
        },
        delWhite(row) {
            this.$confirm(`确认删除 ${row.Serial} ?`, "提示", {
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/white/remove", {
                    serial: row.Serial,
                }).always(() => {
                    this.getWhiteList();
                });
            }).catch(() => {});
        },
        editWhite(row) {
            this.$refs["whiteEditDlg"].show({
                serial: row.Serial,
                password: row.Password,
                user_agent: row.UserAgent,
                remote_ip: row.RemoteIP,
                remote_port: row.RemotePort || "",
                direction: row.Direction || 0,
                cert: row.Cert || "",
                description: row.Description,
            });
        },
        addWhite(row) {
            this.$refs["whiteEditDlg"].show({
                serial: "",
                password: "",
                description: "",
            });
        },
        formatPassword(row, col, cell) {
            if (row.Direction == 1) {
                return "35114 单向";
            } else if (row.Direction == 2) {
                return "35114 双向";
            }
            return cell;
        },
        formatRemoteIP(row, col, cell) {
            if (!cell) {
                return "";
            }
            return cell;
        },
        formatRemotePort(row, col, cell) {
            if (!cell) {
                return "";
            }
            return cell;
        },
    }, //-- methods
    beforeRouteEnter(to, from, next) {
        next(vm => {
            vm.q = to.query.q || "";
            vm.currentPage = parseInt(to.params.page) || 1;
            vm.ready();
        });
    },
    beforeRouteLeave(to, from, next) {
        next();
    },
    beforeRouteUpdate(to, from, next) {
        next();
        this.$nextTick(() => {
            this.q = to.query.q || "";
            this.currentPage = parseInt(to.params.page) || 1;
            this.whiteList = [];
            this.total = 0;
            this.getWhiteList();
        });
    }
};
</script>

<style lang="less">
.opt-group .cell {
    overflow: visible;
}
</style>
