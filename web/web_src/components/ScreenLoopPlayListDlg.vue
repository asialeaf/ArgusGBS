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
                        <label>在线状态</label>
                        <select class="form-control" v-model.trim="online" :disabled="related">
                            <option value="">全部</option>
                            <option value="true">在线</option>
                            <option value="false">离线</option>
                        </select>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>轮播间隔(秒)</label>
                        <el-input-number size="mini" placeholder="轮播间隔 (秒)" style="width:100px;"
                            v-model="interval" :min="10" label="轮播间隔" controls-position="right">
                        </el-input-number>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <div class="checkbox">
                            <el-checkbox style="margin-top:-5px;padding-left:0;" size="small" v-model.trim="related" name="Related">
                                只看已选({{localChannelsAll.length}})
                            </el-checkbox>
                        </div>
                    </div>
                    <div class="form-group form-group-sm pull-right">
                        <button type="button" class="btn btn-primary btn-sm" @click.prevent="startLoopPlay" :disabled="!localChannelsAll || !localChannelsAll.length"><i class="fa fa-play"></i> 开始轮播</button>
                    </div>
                </div>
                <br>
                <el-table :data="related ? localChannels:channels" stripe @sort-change="sortChange" @select="select" @select-all="selectAll" :max-height="500" ref="channelTable" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column type="selection" width="55" fixed :selectable="selectable"></el-table-column>
                    <el-table-column prop="DeviceID" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column prop="ID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                    <!-- <el-table-column prop="DeviceName" label="设备名称" min-width="120" :formatter="formatDeviceName" show-overflow-tooltip></el-table-column> -->
                    <el-table-column prop="Name" label="通道名称" min-width="160" :formatter="formatName" show-overflow-tooltip></el-table-column>
                    <el-table-column min-width="100" label="快照" v-if="!related">
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
                    <el-table-column prop="Status" label="在线状态" min-width="100" v-if="!related">
                        <template slot-scope="props">
                            <span v-if="isDir(props.row)">-</span>
                            <span v-else-if="(props.row.CustomStatus || props.row.Status) == 'ON'" :class="{'text-success': props.row.Status == 'ON', 'text-orange': !!props.row.CustomStatus}">在线</span>
                            <span v-else :class="{'text-orange': !!props.row.CustomStatus}">离线</span>
                        </template>
                    </el-table-column>
                    <!-- <el-table-column prop="SubCount" label="子节点数" min-width="100" v-if="hasAnyRole(serverInfo, userInfo, '管理员')" sortable="custom"></el-table-column> -->
                    <el-table-column prop="Manufacturer" label="厂家" min-width="120" :formatter="formatManufacturer" show-overflow-tooltip></el-table-column>
                </el-table>
                <el-pagination v-if="total > 0 && !related" layout="total,prev,pager,next,sizes" :pageSizes="[10, 20, 50, 100, 200]" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
                <div class="clearfix"></div>
            </div>
            <!-- <div class="modal-footer">
                <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
                <button type="button" class="btn btn-primary" @click.prevent="startLoopPlay" :disabled="!localChannelsAll || !localChannelsAll.length"><i class="fa fa-play"></i> 开始轮播</button>
            </div> -->
        </div> <!-- /.modal-content -->
    </div> <!-- /.modal-dialog -->
</div>
</template>

<script>
import _ from 'lodash'
import $ from 'jquery'
import "jquery-ui/ui/widgets/draggable";

export default {
    props: {
        title: {
            default: '选择轮巡播放通道'
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
            online: "true",
            total: 0,
            pageSize: 10,
            currentPage: 1,
            sort: "",
            order: "",
            loading: false,
            channels: [],
            localData: {
                interval: 10,
                rows: {},
            },
            related: false,
            localChannels: [],
            localChannelsAll: [],
            interval: 10,
        };
    },
    watch: {
        q: function (newVal, oldVal) {
            this.doDelaySearch();
        },
        online: function(newVal, oldVal) {
            this.doSearch();
        },
        currentPage: function (newVal, oldVal) {
            this.doSearch(newVal);
        },
        pageSize: function (newVal, oldVal) {
            this.doSearch();
        },
        related: function (newVal, oldVal) {
            if (newVal) {
                this.loadLocalChannels();
            } else {
                this.doSearch();
            }
        },
        interval: function (newVal, oldVal) {
            this.localData.interval = newVal;
            this.setLocalData(true);
        },
    },
    mounted() {
        this.loadLocalData();
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
            // this.reset();
            this.$emit("hide");
        })
    },
    methods: {
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order == "ascending" ? "asc" : "desc";
            this.getChannels();
        },
        doSearch(page = 1) {
            this.currentPage = page;
            if (this.related) {
                this.loadLocalChannels();
            } else {
                this.getChannels();
            }
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
        getChannels() {
            this.loading = true;
            $.get("/api/v1/device/channellist", {
                q: this.q,
                start: (this.currentPage - 1) * this.pageSize,
                limit: this.pageSize,
                channel_type: "device",
                online: this.online,
                sort: this.sort,
                order: this.order
            }).then(ret => {
                this.total = ret.ChannelCount;
                this.channels = ret.ChannelList || [];
                this.$nextTick(() => {
                    this.channels.forEach(row => {
                        if (this.localData.rows[row.DeviceID + "_" + row.ID]) {
                            this.localData.rows[row.DeviceID + "_" + row.ID] = row;
                            this.$refs["channelTable"].toggleRowSelection(row);
                        }
                    })
                    this.setLocalData(false);
                })
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
            });
        },
        // rowClick(row, event, column) {
        //     if (!row) {
        //         return
        //     }
        //     if (this.localData.rows[row.DeviceID + "_" + row.ID]) {
        //         delete this.localData.rows[row.DeviceID + "_" + row.ID];
        //     } else {
        //         this.localData.rows[row.DeviceID + "_" + row.ID] = row;
        //     }
        //     this.$refs["channelTable"].toggleRowSelection(row);
        //     this.setLocalData(false);
        // },
        // rowStyle({row, rowIndex}) {
        //     if ((row.CustomStatus || row.Status) === "ON") {
        //         return "cursor:pointer";
        //     }
        //     return "";
        // },
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
            this.channels = [];
            this.q = "";
            this.online = "true";
            this.currentPage = 1;
            this.pageSize = 10;
        },
        show() {
            $(this.$el).modal("show");
            if (this.related) {
                this.loadLocalData()
            } else {
                this.getChannels();
            }
        },
        hide() {
            $(this.$el).modal("hide");
        },
        selectable(row, index) {
            return true;
        },
        select(selection, row) {
            var r = this.localData.rows[row.DeviceID + "_" + row.ID] || "";
            var idx = selection.indexOf(row);
            if (idx >= 0) {
                if (!r) {
                    this.localData.rows[row.DeviceID + "_" + row.ID] = row;
                }
            } else {
                if (r) {
                    delete this.localData.rows[row.DeviceID + "_" + row.ID];
                }
            }
            this.setLocalData();
        },
        selectAll(selection) {
            if (selection.length) {
                this.selection = selection;
                for (var row of selection) {
                    var r = this.localData.rows[row.DeviceID + "_" + row.ID] || "";
                    if (!r) {
                        this.localData.rows[row.DeviceID + "_" + row.ID] = row;
                    }
                }
            } else {
                if (this.related) {
                    for (var row of this.localChannels) {
                        var r = this.localData.rows[row.DeviceID + "_" + row.ID] || "";
                        if (r) {
                            delete this.localData.rows[row.DeviceID + "_" + row.ID];
                        }
                    }
                } else {
                    for (var row of this.channels) {
                        var r = this.localData.rows[row.DeviceID + "_" + row.ID] || "";
                        if (r) {
                            delete this.localData.rows[row.DeviceID + "_" + row.ID];
                        }
                    }
                }
            }
            this.setLocalData();
        },
        setLocalData(updateSet) {
            localStorage['livegbs_screen_loop_play'] = JSON.stringify(this.localData);
            if (!updateSet) {
                this.loadLocalChannels();
            }
        },
        loadLocalData() {
            if (localStorage['livegbs_screen_loop_play']) {
                this.localData = JSON.parse(localStorage['livegbs_screen_loop_play']);
                this.interval = this.localData.interval;
                this.loadLocalChannels();
            }
        },
        async loadLocalChannels() {
            var localRows = [];
            var localRowsAll = [];
            for (var key in this.localData.rows) {
                var r = this.localData.rows[key];
                if(!r) {
                    delete this.localData.rows[key];
                    continue
                }
                if(!this.hasAllChannel(this.serverInfo, this.userInfo)) {
                    var ok = await this.hasChannel(r.DeviceID, r.ID);
                    if(!ok) {
                        delete this.localData.rows[key];
                        continue
                    }
                }
                localRowsAll.push(r);
                if(this.q && (r.ID + r.DeviceID + r.Name + "").indexOf(this.q) < 0) {
                    continue
                }
                localRows.push(r);
            }
            this.localChannelsAll = localRowsAll;
            this.localChannels = localRows;
            if (this.related) {
                this.$nextTick(() => {
                    this.localChannels.forEach(row => {
                        this.$refs["channelTable"].toggleRowSelection(row);
                    })
                })
            }
        },
        hasChannel(serial, code) {
            return new Promise((resolve, reject) => {
                $.ajax({
                    method: "GET",
                    url: "/api/v1/device/channelinfo",
                    global: false,
                    data: {
                        serial: serial,
                        code: code,
                    }
                }).then(data => {
                    resolve(true);
                }).fail(xhr => {
                    resolve(false);
                })
            })
        },
        startLoopPlay() {
            if (!this.localChannelsAll || !this.localChannelsAll.length) {
                this.$message({
                    type: "warning",
                    message: "请选择需要轮巡播放的通道"
                })
                return
            }
            this.$emit("loopPlay", this.localChannelsAll, this.interval);
            this.hide();
        },
    }, //-- methods
};
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px) {
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px) {
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }
</style>
