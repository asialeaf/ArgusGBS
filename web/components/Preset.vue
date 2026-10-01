<template>
<div>
    <el-table size="mini" :data="presets" stripe :max-height="235" ref="presetsTable" v-loading="loading" element-loading-text="获取中...">
        <el-table-column prop="PresetID" label="序号" min-width="40" show-overflow-tooltip></el-table-column>
        <el-table-column prop="PresetName" label="备注" min-width="80" show-overflow-tooltip>
            <template slot-scope="props">
                <a role="button" @click.prevent="setCustomName(props.row, true, $event)" v-if="!props.row.Editing">{{props.row.PresetName || '-'}}</a>
                <input type="text" style="width:80px;padding:2px;line-height:100%;"
                    @keydown.esc.stop.prevent="setCustomName(props.row, false, $event)"
                    @keydown.enter.stop.prevent="setCustomName(props.row, false, $event)"
                    @blur="setCustomName(props.row, false, $event)" :value="props.row.PresetName"
                    v-else v-focus>
            </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" width="110">
            <template slot-scope="scope">
                <span title="调用预置位" style="cursor:pointer;color:#00acd6;"
                    @click.prevent="gotoPreet(scope.row)" v-if="scope.row.PresetEnable && !scope.row.Going">
                    调用
                </span>
                <span v-else-if="scope.row.Going"> <i class="fa fa-spinner fa-spin"></i> </span>
                <span title="设置预置位" style="cursor:pointer;color:#00acd6;"
                    @click.prevent="setPreset(scope.row)" v-if="!scope.row.Setting">
                    设置
                </span>
                <span v-else> <i class="fa fa-spinner fa-spin"></i> </span>
                <span title="删除预置位" :style="scope.row.Confirm ? 'cursor:pointer;color:#e08e0b;':'cursor:pointer;color:#dd4b39;'"
                    @click.prevent="removePreset(scope.row)" v-if="scope.row.PresetEnable">
                    {{scope.row.Confirm ? '确认?' : '删除'}}
                </span>
            </template>
        </el-table-column>
    </el-table>
</div>
</template>

<script>
export default {
    props: {
        serial: {
            type: String,
            default: ''
        },
        code: {
            type: String,
            default: ''
        },
        activeName: {
            type: String,
            default: ''
        },
    },
    data() {
        return {
            presets: [],
            loading: false,
        }
    },
    watch: {
        activeName: function (newVal, oldVal) {
            if (newVal == "preset-info") {
                this.getPresets();
            }
        },
    },
    mounted() {},
    directives: {
        focus: {
            inserted: function (el) {
                el.focus();
                el.select();
            }
        }
    },
    methods: {
        getPresets() {
            this.loading = true;
            $.ajax({
                method: "GET", // type: "GET", is also ok, method since jquery 1.9
                url: "/api/v1/device/fetchpreset",
                global: false,
                data: {
                    serial: this.serial,
                    code: this.code,
                    // timeout: 3, // some dahua device response presets take more than 3 secs
                    timeout: 5,
                }
            }).then(ret => {
                this.presets = ret.PresetItemList || [];
            }).fail(xhr => {
                if(xhr) {
                    console.log(`fetchpreset[${this.serial}:${this.code}] ajax error: ${xhr.status} ${xhr.responseText}`);
                    if(xhr.status == 401) {
                        this.$nextTick(() => {
                            location.href = `/login?r=${encodeURIComponent(window.location.href)}`;
                        })
                    }
                }
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
            });
        },
        setCustomName(row, editing, event) {
            var _editing = row.Editing;
            this.$set(row, "Editing", editing);
            if(_editing && !editing) {
                var val = event.target.value.trim();
                if(row.PresetName != val) {
                    $.post('/api/v1/device/setpresetname', {
                        serial: this.serial,
                        code: this.code,
                        preset: row.PresetID,
                        name: val,
                    }).always(() => {
                        this.getPresets();
                    })
                }
            }
        },
        setPreset(row) {
            this.$set(row, "Setting", true);
            $.post("/api/v1/control/preset", {
                serial: this.serial,
                code: this.code,
                command: "set",
                preset: row.PresetID,
                name: row.PresetName,
            }).always(() => {
                this.$delete(row, "Setting");
                this.getPresets();
            });
        },
        removePreset(row) {
            if (row.Confirm) {
                $.post("/api/v1/control/preset", {
                    serial: this.serial,
                    code: this.code,
                    command: "remove",
                    preset: row.PresetID,
                }).always(() => {
                    this.$delete(row, "Confirm");
                    this.getPresets();
                });
            } else {
                this.$set(row, "Confirm", true);
                setTimeout(() => {
                    this.$delete(row, "Confirm");
                }, 3000);
            }
        },
        gotoPreet(row) {
            this.$set(row, "Going", true);
            $.post("/api/v1/control/preset", {
                serial: this.serial,
                code: this.code,
                command: "goto",
                preset: row.PresetID,
            }).always(() => {
                this.$delete(row, "Going");
            })
        },
    }
}
</script>
