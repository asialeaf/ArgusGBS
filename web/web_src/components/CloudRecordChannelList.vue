<template>
<div class="box box-primary ids">
    <div class="box-header">
        <h4 class="text-primary text-center">云端录像通道列表</h4>
    </div>
    <div class="box-body">
        <div class="form-inline" autocomplete="off" spellcheck="false">
            <div class="form-group form-group-sm" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')">
                <div class="input-group input-group-sm">
                    <router-link :to="`/cloudrecord/plans/1`" class="btn btn-sm btn-primary">
                        <i class="fa fa-calendar"></i> 录像计划
                    </router-link>
                </div>
            </div>
            <span class="hidden-xs" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')">&nbsp;&nbsp;</span>
            <div class="form-group form-group-sm">
                <label>搜索</label>
                <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
            </div>
        </div>
        <br>
        <div class="clearfix"></div>
        <el-table :data="pageData" stripe :default-sort="{prop: 'serial', order: 'ascending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
            <el-table-column min-width="200" label="设备国标编号" prop="serial" show-overflow-tooltip sortable="custom"></el-table-column>
            <el-table-column min-width="200" label="通道国标编号" prop="code" show-overflow-tooltip sortable="custom"></el-table-column>
            <el-table-column min-width="160" label="操作" :fixed="isMobile() ? false : 'right'">
                <template slot-scope="scope">
                    <div class="btn-group btn-group-xs">
                        <router-link :to="`/cloudrecord/timeview/${scope.row.serial}/${scope.row.code}`" class="btn btn-info">
                            <i class="fa fa-file-video-o"></i> 查看录像
                        </router-link>
                        <a role="button" class="btn btn-danger" @click.prevent="remove(scope.row)" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                            <i class="fa fa-remove"></i> 删除
                        </a>
                    </div>
                </template>
            </el-table-column>
            <el-table-column min-width="160" label="通道名称" prop="name" :formatter="formatName" show-overflow-tooltip></el-table-column>
            <el-table-column min-width="100" label="快照">
                <template slot-scope="props">
                    <el-popover :open-delay="1000" :close-delay="10" placement="left" :title="props.row.code" width="400" trigger="hover">
                        <img onerror='this.src="/images/default_snap.png";' style="width:100%;height:100%;" :src="props.row.snap_url">
                        <img onerror='this.src="/images/default_snap.png";' style="height:30px;width:50px;" slot="reference" :src="props.row.snap_url">
                    </el-popover>
                </template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" min-width="160" sortable="custom"></el-table-column>
            <el-table-column prop="created_at" label="创建时间" min-width="160" sortable="custom"></el-table-column>
        </el-table>
    </div>
    <div class="box-footer clearfix" v-if="total > 0">
        <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="page"></el-pagination>
    </div>
</div>
</template>

<script>
import { mapState } from 'vuex'
import _ from 'lodash'

export default {
    data() {
        return {
            pageSize: 10,
            q: "",
            page: 1,
            pageData: [],
            total: 0,
            sort: '',
            order: '',
            loading: false,
        }
    },
    mounted() {
        // this.$refs['q'].focus();
        // this.load(); // load when sort change
    },
    computed: {
        ...mapState(['serverInfo', 'userInfo'])
    },
    methods: {
        ready(){
            this.$watch('q', function() {
                this.doDelaySearch();
            });
            this.$watch('page', function(val) {
                this.doSearch(val);
            });
            // this.load(); // load when sort change
        },
        load(global = true) {
            if(global) {
                this.loading = true;
            }
            $.ajax({
                method: "GET",
                url: '/api/v1/cloudrecord/querychannels',
                global: global,
                data: {
                    start: (this.page - 1) * this.pageSize,
                    limit: this.pageSize,
                    q: this.q,
                    sort: this.sort,
                    order: this.order
                }
            }).then(data => {
                this.pageData = data.rows;
                this.total = data.total;
            }).always(() => {
                if(global) {
                    this.$nextTick(() => {
                        this.loading = false;
                    })
                }
            })
        },
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order == "ascending" ? "asc" : "desc";
            this.load();
        },
        selectDevice(id) {
            this.$emit("select-device", id);
        },
        doSearch(page = 1) {
            var query = {};
            if (this.q) query["q"] = this.q;
            this.$router.replace({
                path: `/cloudrecord/${page}`,
                query: query
            });
        },
        doDelaySearch: _.debounce(function () {
            this.doSearch();
        }, 800),
        formatName(row, col, cell) {
            var devName = row.device_custom_name || row.device_name || "";
            var chName = row.custom_name || row.name || "";
            if(devName && devName != chName) {
                if(!chName) {
                    chName = devName;
                } else {
                    chName = `${chName}@${devName}`;
                }
            }
            return chName || "-";
        },
        async remove(row) {
            this.$confirm(`删除 ${row.custom_name || row.name || row.code || '-'} , 会将该通道所有录像删除，确认继续?`, '提示', {
                lockScroll: false,
            }).then(() => {
                $.post('/api/v1/cloudrecord/removechannel', {
                    serial: row.serial,
                    code: row.code,
                }).always(() => {
                    this.load();
                })
            }).catch(() => {})
        }
    },
    beforeRouteEnter(to, from, next) {
        next(vm => {
            vm.q = to.query.q || "";
            vm.page = parseInt(to.params.page) || 1;
            vm.ready();
        })
    },
    beforeRouteUpdate(to, from, next) {
        next();
        this.$nextTick(() => {
            this.q = to.query.q || "";
            this.page = parseInt(to.params.page) || 1;
            this.pageData = [];
            this.total = 0;
            this.load();
        })
    }
}
</script>