<template>
<div :class="['modal', {fade: fade}]" data-backdrop="static" data-disable="false" data-keyboard="true" tabindex="-1">
    <div :class="['modal-dialog', size]">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                    <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title text-center text-primary"><span>{{title}}</span></h4>
            </div>
            <div class="modal-body">
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group form-group-sm">
                        <label>搜索</label>
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>报警订阅状态</label>
                        <select class="form-control" v-model.trim="subscribed">
                            <option value="">全部</option>
                            <option value="true">已订阅</option>
                            <option value="false">未订阅</option>
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
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <div class="checkbox" style="user-select:none;">
                            <el-checkbox style="margin-top:-5px;padding-left:0;" size="small" v-model.trim="related" name="Related">
                                只看{{reverse ? '未选' : '已选'}}({{relateCnt}})
                            </el-checkbox>
                            <span>
                                <a role="button" @click="toggleReverse">
                                    <i class="fa fa-sort"></i>
                                </a>
                            </span>
                        </div>
                    </div>
                </div>
                <br>
                <el-table :data="channels" stripe @sort-change="sortChange" @select="select" @select-all="selectAll" :max-height="500" ref="channelTable" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column type="selection" width="55" fixed :selectable="selectable"></el-table-column>
                    <el-table-column prop="DeviceID" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column prop="ID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                    <!-- <el-table-column prop="DeviceName" label="设备名称" min-width="120" :formatter="formatDeviceName" show-overflow-tooltip></el-table-column> -->
                    <el-table-column prop="Name" label="通道名称" min-width="160" :formatter="formatName" show-overflow-tooltip></el-table-column>
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
                    <el-table-column prop="AlarmSubscribe" label="报警订阅状态" min-width="140" :formatter="formatAlarmSubscribe" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Status" label="在线状态" min-width="100">
                        <template slot-scope="props">
                            <span v-if="isDir(props.row)">-</span>
                            <span v-else-if="props.row.DeviceOnline && (props.row.CustomStatus || props.row.Status) == 'ON'" :class="{'text-success': props.row.Status == 'ON', 'text-orange': !!props.row.CustomStatus}">在线</span>
                            <span v-else :class="{'text-orange': !!props.row.CustomStatus}">离线</span>
                        </template>
                    </el-table-column>
                    <el-table-column prop="Manufacturer" label="厂家" min-width="120" :formatter="formatManufacturer" show-overflow-tooltip></el-table-column>
                </el-table>
                <el-pagination v-if="total > 0" layout="total,prev,pager,next,sizes" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
                <div class="clearfix"></div>
            </div>
            <!-- <div class="modal-footer">
                <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
            </div> -->
        </div>
    </div>
</div>
</template>

<script>
import 'jquery-ui/ui/widgets/draggable';
import $ from 'jquery';
import _ from "lodash";

export default {
    props: {
        title: {
            default: ''
        },
        size: {
            type: String,
            default: 'modal-lgg'
        },
        fade: {
            type: Boolean,
            default: false
        },
		serverInfo: {
			type: Object,
			default: () => {}
		},
		userInfo: {
			type: Object,
			default: () => null
		}
    },
    data() {
        return {
            q: "",
            subscribed: "",
            online: "",
            total: 0,
            relateCnt: 0,
            pageSize: 10,
            currentPage: 1,
            sort: "",
            order: "",
            related: false,
            reverse: false,
            loading: false,
            channels: [],
            globalDeviceAlarmSubscribeInterval: 0,
            selection: [],
            id: '', //外部关联id
        }
    },
    watch: {
        q: function (newVal, oldVal) {
            this.doDelaySearch();
        },
        related: function(newVal, oldVal) {
            this.doSearch();
        },
        reverse: function(newVal, oldVal) {
            this.doSearch();
        },
        online: function(newVal, oldVal) {
            this.doSearch();
        },
        subscribed: function(newVal, oldVal) {
            this.doSearch();
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
            this.$emit("show");
        }).on("hidden.bs.modal", () => {
            this.errors.clear();
            this.reset();
            this.$emit("hide");
        })
    },
    directives: {
        focus: {
            inserted: function (el) {
                el.focus();
                el.select();
            }
        }
    },
    methods: {
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order == "ascending" ? "asc" : "desc";
            this.getChannels();
        },
        select(selection, row) {
            var idx = selection.indexOf(row);
            if (idx >= 0) {
                $.post("/api/v1/alarm/plan/savechannels", {
                    id: this.id,
                    channels: [`${row.DeviceID}:${row.ID}`]
                }).always(() => {
                    this.getChannels();
                })
            } else {
                $.post("/api/v1/alarm/plan/removechannels", {
                    id: this.id,
                    channels: [`${row.DeviceID}:${row.ID}`]
                }).always(() => {
                    this.getChannels();
                })
            }
        },
        selectAll(selection) {
            var keys = [];
            if (selection.length) {
                for (var row of selection) {
                    var idx = this.selection.indexOf(row);
                    if (idx < 0) {
                        keys.push(`${row.DeviceID}:${row.ID}`)
                    }
                }
                $.post("/api/v1/alarm/plan/savechannels", {
                    id: this.id,
                    channels: keys,
                }).always(() => {
                    this.getChannels();
                })
            } else {
                for (var row of this.selection) {
                    keys.push(`${row.DeviceID}:${row.ID}`)
                }
                $.post("/api/v1/alarm/plan/removechannels", {
                    id: this.id,
                    channels: keys,
                }).always(() => {
                    this.getChannels();
                })
            }
        },
        doSearch(page = 1) {
            this.currentPage = page;
            this.getChannels();
        },
        doDelaySearch: _.debounce(function () {
            this.doSearch();
        }, 800),
        formatName(row, col, cell) {
            var devName = row.DeviceCustomName || row.DeviceName || "";
            var chName = row.CustomName || row.Name || "";
            if(devName && devName != chName) {
                if(!chName) {
                    chName = devName;
                } else {
                    chName = `${chName}@${devName}`;
                }
            }
            return chName || "-";
        },
        formatDeviceName(row, col, cell) {
            return row.DeviceCustomName || row.DeviceName || "-";
        },
        formatChannelName(row, col, cell) {
            return row.CustomName || row.Name || "-";
        },
        formatManufacturer(row, col, cell) {
            if (cell) return cell;
            return "-";
        },
        formatAlarmSubscribe(row, col, cell) {
            if (row.SubscribeInterval < 0) {
                return "未订阅";
            } else if(row.SubscribeInterval == 0) {
                if(this.globalDeviceAlarmSubscribeInterval <= 0) {
                    return "未订阅";
                } else {
                    return "已全局订阅";
                }
            } else if(row.SubscribeInterval > 0 && !row.AlarmSubscribe) {
                return "未订阅";
            }
            return "已订阅";
        },
        selectable(row, index) {
            return true;
        },
        getChannels() {
            if (!this.id) return;
            this.loading = true;
            $.get("/api/v1/alarm/plan/channellist", {
                id: this.id,
                q: this.q,
                subscribed: this.subscribed,
                online: this.online,
                start: (this.currentPage - 1) * this.pageSize,
                limit: this.pageSize,
                related: this.related,
                reverse: this.reverse,
                sort: this.sort,
                order: this.order
            }).then(ret => {
                this.$refs["channelTable"].clearSelection();
                this.globalDeviceAlarmSubscribeInterval = ret.GlobalDeviceAlarmSubscribeInterval || 0;
                this.total = ret.ChannelCount;
                this.relateCnt = ret.ChannelRelateCount;
                this.channels = ret.ChannelList || [];
                this.selection = [];
                this.$nextTick(() => {
                    this.channels.forEach(row => {
                        var sel = row.PlanID != "";
                        this.$refs["channelTable"].toggleRowSelection(row, sel);
                        if (sel) {
                            this.selection.push(row);
                        }
                    })
                })
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
            });
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
        reset() {
            this.id = '';
            this.$refs["channelTable"].clearSelection();
            this.channels = [];
            this.selection = [];
            this.q = "";
            this.subscribed = "";
            this.online = "";
            this.related = false;
            this.reverse = false;
            this.total = 0;
            this.relateCnt = 0;
            this.currentPage = 1;
            this.pageSize = 10;
        },
        show(id) {
            this.id = id;
            $(this.$el).modal("show");
            this.getChannels();
        },
        hide() {
            $(this.$el).modal("hide");
        },
        toggleReverse() {
            this.reverse = !this.reverse;
        },
    }, //-- methods
}
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px){
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px){
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }
</style>
