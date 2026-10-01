<template>
<div :class="['modal', {fade: fade}]" data-backdrop="static" data-disable="false" data-keyboard="true" tabindex="-1">
    <div :class="['modal-dialog', size]">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                    <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title text-center text-primary"><span>{{title||'录像通道'}}</span></h4>
            </div>
            <div class="modal-body">
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group form-group-sm">
                        <label>搜索</label>
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                    </div>
                    <div class="form-group form-group-sm pull-right">
                        <div class="btn-group btn-group-sm">
                            <button type="button" class="btn btn-primary" @click.prevent="getChannels"><i class="fa fa-refresh"></i> 刷新</button>
                        </div>
                    </div>
                </div>
                <br>
                <div class="clearfix"></div>
                <el-table :data="channels" stripe :default-sort="{prop: 'serial', order: 'ascending'}" @sort-change="sortChange" :max-height="500" ref="channelTable" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column min-width="200" label="设备国标编号" prop="serial" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column min-width="200" label="通道国标编号" prop="code" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column min-width="200" label="操作" :fixed="isMobile() ? false : 'right'">
                        <template slot-scope="scope">
                            <div class="btn-group btn-group-xs">
                                <a role="button" class="btn btn-primary btn-xs" @click.prevent="channelChange(scope.row.serial, scope.row.code)">
                                    <i class="fa fa-file-video-o"></i> 查看录像
                                </a>
                            </div>
                            <el-tag size="mini" v-if="scope.row.serial == serial && scope.row.code == code">当前通道</el-tag>
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
                <el-pagination v-if="total > 0" layout="total,prev,pager,next,sizes" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
                <div class="clearfix"></div>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import 'jquery-ui/ui/widgets/draggable';
import $ from 'jquery';
import _ from "lodash";
import { mapState } from "vuex";

export default {
    props: {
        title: {
            default: '录像通道'
        },
        size: {
            type: String,
            default: 'modal-lgg'
        },
        fade: {
            type: Boolean,
            default: false
        }
    },
    data() {
        return {
            q: "",
            total: 0,
            pageSize: 10,
            currentPage: 1,
            sort: 'id',
            order: 'ascending',
            loading: false,
            shown: false,
            channels: [],
            day: "",
            viewType: "timeview",
            serial: "",
            code: "",
        }
    },
    computed: {
        ...mapState(['userInfo', 'serverInfo']),
    },
    watch: {
        q: function (newVal, oldVal) {
            this.doDelaySearch();
        },
        currentPage: function (newVal, oldVal) {
            this.doSearch(newVal);
        },
        pageSize: function (newVal, oldVal) {
            this.doSearch();
        }
    },
    mounted() {
        $(this.$el).find('.modal-content').draggable({
            handle: '.modal-header',
            cancel: ".modal-title span",
            addClasses: false,
            containment: 'document',
            delay: 100,
            opacity: 0.5
        });
        $(this.$el).on("shown.bs.modal", () => {
            this.shown = true;
            this.$emit("show");
            this.getChannels();
        }).on("hidden.bs.modal", () => {
            this.shown = false;
            this.errors.clear();
            this.reset();
            this.$emit("hide");
        })
    },
    methods: {
        channelChange(serial, code) {
            this.$router.replace(`/cloudrecord/${this.viewType}/${serial}/${code}/${this.day}`);
            this.hide();
        },
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order;
            this.getChannels();
        },
        doSearch(page = 1) {
            this.currentPage = page;
            this.getChannels();
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
        getChannels(e) {
            if (!this.shown) return;
            this.loading = true;
            $.get("/api/v1/cloudrecord/querychannels", {
                start: (this.currentPage - 1) * this.pageSize,
                limit: this.pageSize,
                q: this.q,
                sort: this.sort,
                order: this.order
            }).then(data => {
                this.total = data.total || 0;
                this.channels = data.rows || [];
                if (e && e.target) {
                    this.$message({
                        type: "success",
                        message: "刷新成功！",
                    });
                }
            }).always(() => {
                this.loading = false;
            });
        },
        reset() {
            this.channels = [];
            this.total = 0;
            this.day = "";
            this.viewType = "timeview";
            this.serial = "";
            this.code = "";
            // this.q = "";
            // this.currentPage = 1;
            // this.pageSize = 10;
        },
        show(day = '', viewType = 'timeview', serial = '', code = '') {
            this.day = day;
            this.viewType = viewType;
            this.serial = serial;
            this.code = code;
            $(this.$el).modal("show");
        },
        hide() {
            $(this.$el).modal("hide");
        },
    }
}
</script>

<style lang="less" scoped>
.modal-content {
    overflow: hidden;
}

@media screen and(min-width: 992px) {
    .modal-dialog.modal-lgg {
        width: 80%;
    }
}

@media screen and(min-width: 1200px) {
    .modal-dialog.modal-lgg {
        width: 1200px;
    }
}
</style>
