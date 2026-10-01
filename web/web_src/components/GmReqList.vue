<template>
<div>
    <div class="box box-primary">
        <div class="box-header">
            <h4 class="text-primary text-center">证书请求列表</h4>
        </div>
        <div class="box-body">
            <div class="form-inline" autocomplete="off" spellcheck="false">
                <div class="form-group form-group-sm">
                    <div class="input-group input-group-sm">
                        <el-upload action="/api/v1/gm/req/import"
                            style="display:inline-block;margin-right:7px;" :show-file-list="false"
                            :on-success="onUploadReqCertSuccess" :on-error="onUploadReqCertError" :on-progress="onUploadReqCertProgress">
                            <button type="button" class="btn btn-sm btn-primary" title="上传证书请求"><i class="fa fa-upload"></i> 上传证书请求</button>
                        </el-upload>
                    </div>
                </div>
                <span class="hidden-xs">&nbsp;&nbsp;</span>
                <div class="form-group form-group-sm">
                    <label>搜索</label>
                    <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                </div>
            </div>
            <br>
            <div class="clearfix"></div>
            <el-table :data="reqList" stripe :default-sort="{prop: 'CreatedAt', order: 'descending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                <el-table-column prop="Serial" label="序列号" min-width="280" show-overflow-tooltip sortable="custom" :formatter="formatCell"></el-table-column>
                <el-table-column label="操作" min-width="150" :fixed="isMobile() ? false : 'right'">
                    <template slot-scope="props">
                        <div class="btn-group btn-group-xs">
                            <button type="button" class="btn btn-primary" @click.prevent="signReq(props.row, props.column)" v-if="userInfo">
                                <i class="fa fa-download"></i> 签发证书
                            </button>
                            <button type="button" class="btn btn-danger" @click.prevent="removeReq(props.row, props.column)" v-if="userInfo">
                                <i class="fa fa-remove"></i> 删除
                            </button>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="CommonName" label="签发给" min-width="200" show-overflow-tooltip sortable="custom" :formatter="formatCommonName"></el-table-column>
                <el-table-column prop="Description" label="备注" min-width="200" show-overflow-tooltip>
                    <template slot-scope="props">
                        <a role="button" @click.prevent="remarkReq(props.row, props.column)" title="编辑备注">
                            <i class="fa fa-edit"></i>
                        </a>
                        {{props.row.Description||'-'}}
                    </template>
                </el-table-column>
                <el-table-column prop="CreatedAt" label="创建时间" min-width="160" sortable="custom"></el-table-column>
            </el-table>
        </div>
        <div class="box-footer" v-if="total > 0">
            <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
        </div>
        <GmReqSignDlg ref="signDlg"></GmReqSignDlg>
        <GmReqRemarkDlg ref="remarkDlg" @submit="getReqList"></GmReqRemarkDlg>
    </div>
</div>
</template>

<script>
import GmReqSignDlg from "components/GmReqSignDlg.vue";
import GmReqRemarkDlg from "components/GmReqRemarkDlg.vue";
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
            sort: "CreatedAt",
            order: "desc",
            loading: false,
            bUploading: false,
            reqList: [],
        };
    },
    components: {
        GmReqSignDlg, GmReqRemarkDlg
    },
    computed: {
        ...mapState(['userInfo', 'serverInfo']),
    },
    mounted() {
        // this.getReqList();
    },
    beforeDestroy() {
    },
    methods: {
        ready(){
            this.$watch('q', function(newVal, oldVal) {
                this.doDelaySearch();
            });
            this.$watch('currentPage', function(newVal, oldVal) {
                this.doSearch(newVal);
            });
            // this.getReqList();
        },
        doSearch(page = 1) {
            var query = {};
            if (this.q) query["q"] = this.q;
            this.$router.replace({
                path: `/gm/req/${page}`,
                query: query
            });
        },
        doDelaySearch: _.debounce(function () {
            this.doSearch();
        }, 800),
        getReqList() {
            this.loading = true;
            $.get("/api/v1/gm/req/list", {
                q: this.q,
                start: (this.currentPage - 1) * this.pageSize,
                limit: this.pageSize,
                sort: this.sort,
                order: this.order
            }).then(ret => {
                this.total = ret.GMReqCount;
                this.reqList = ret.GMReqList;
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
            });
        },
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order == "ascending" ? "asc" : "desc";
            this.getReqList();
        },
        formatCell(row, col, cell) {
            return row[col.property] || "-";
        },
        formatCommonName(row, col, cell) {
            var ret = cell || "-";
            var list = ret.split("_");
            if(list.length == 2) {
                ret = list[0] || "-";
            }
            return ret;
        },
        removeReq(row, col) {
            this.$confirm(`确认删除 ${row.Serial||this.formatCommonName(row, col, row.CommonName)} ?`, "提示", {
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/gm/req/remove", {
                    id: row.ID,
                }).always(() => {
                    this.getReqList();
                });
            }).catch(() => {});
        },
        signReq(row, col) {
            this.$refs["signDlg"].show(row.ID, this.formatCommonName(row, col, row.CommonName));
        },
        remarkReq(row, col) {
            this.$refs["remarkDlg"].show({
                id: row.ID,
                serial: row.Serial||"-",
                common_name: this.formatCommonName(row, col, row.CommonName),
                description: row.Description,
            });
        },
        onUploadReqCertSuccess(res, file, fileList) {
            this.bUploading = false;
            this.getReqList();
            this.$message({
                type: "success",
                message: "上传证书请求成功！"
            });
        },
        onUploadReqCertProgress(evt, file, fileList) {
            this.bUploading = true;
        },
        onUploadReqCertError(err, file, fileList) {
            this.bUploading = false;
            if (err) {
                this.$message({
                    type: "error",
                    message: err + ""
                })
            }
        },
    },
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
            this.reqList = [];
            this.total = 0;
            this.getReqList();
        });
    }
};
</script>
