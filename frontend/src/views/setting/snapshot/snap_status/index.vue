<template>
    <el-dialog
        v-model="dialogVisible"
        @close="onClose"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="50%"
    >
        <template #header>
            <div class="card-header">
                <span>{{ $t('setting.status') }}</span>
            </div>
        </template>
        <div v-loading="loading">
            <el-alert :type="loadStatus(status.panelInfo)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.panelInfo)" link>{{ $t('setting.panelInfo') }}</el-button>
                    <div v-if="showErrorMsg(status.panelInfo)" class="top-margin">
                        <span class="err-message">{{ status.panelInfo }}</span>
                    </div>
                </template>
            </el-alert>
            <el-alert :type="loadStatus(status.panel)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.panel)" link>{{ $t('setting.panelBin') }}</el-button>
                    <div v-if="showErrorMsg(status.panel)" class="top-margin">
                        <span class="err-message">{{ status.panel }}</span>
                    </div>
                </template>
            </el-alert>
            <el-alert :type="loadStatus(status.daemonJson)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.daemonJson)" link>{{ $t('setting.daemonJson') }}</el-button>
                    <div v-if="showErrorMsg(status.daemonJson)" class="top-margin">
                        <span class="err-message">{{ status.daemonJson }}</span>
                    </div>
                </template>
            </el-alert>
            <el-alert :type="loadStatus(status.appData)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.appData)" link>{{ $t('setting.appData') }}</el-button>
                    <div v-if="showErrorMsg(status.appData)" class="top-margin">
                        <span class="err-message">{{ status.appData }}</span>
                    </div>
                </template>
            </el-alert>
            <el-alert :type="loadStatus(status.backupData)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.backupData)" link>{{ $t('setting.backupData') }}</el-button>
                    <div v-if="showErrorMsg(status.backupData)" class="top-margin">
                        <span class="err-message">{{ status.backupData }}</span>
                    </div>
                </template>
            </el-alert>
            <el-alert :type="loadStatus(status.panelData)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.panelData)" link>{{ $t('setting.panelData') }}</el-button>
                    <div v-if="showErrorMsg(status.panelData)" class="top-margin">
                        <span class="err-message">{{ status.panelData }}</span>
                    </div>
                </template>
            </el-alert>
            <el-alert :type="loadStatus(status.compress)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.compress)" link>
                        {{ $t('setting.compress') }} {{ status.size }}
                    </el-button>
                    <div v-if="showErrorMsg(status.compress)" class="top-margin">
                        <span class="err-message">{{ status.compress }}</span>
                    </div>
                </template>
            </el-alert>
            <el-alert :type="loadStatus(status.upload)" :closable="false">
                <template #title>
                    <el-button :icon="loadIcon(status.upload)" link>
                        {{ $t('setting.upload') }}
                    </el-button>
                    <div v-if="showErrorMsg(status.upload)" class="top-margin">
                        <span class="err-message">{{ status.upload }}</span>
                    </div>
                </template>
            </el-alert>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="onClose">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button v-if="showRetry()" @click="onRetry">
                    {{ $t('commons.button.retry') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { Setting } from '@/api/interface/setting';
import { loadSnapStatus, snapshotCreate } from '@/api/modules/setting';
import { nextTick, onBeforeUnmount, reactive, ref } from 'vue';

const status = reactive<Setting.SnapshotStatus>({
    panel: '',
    panelInfo: '',
    daemonJson: '',
    appData: '',
    panelData: '',
    backupData: '',

    compress: '',
    size: '',
    upload: '',
});

const dialogVisible = ref(false);

const loading = ref();
const snapID = ref();
const snapFrom = ref();
const snapDescription = ref();

let timer: NodeJS.Timer | null = null;

interface DialogProps {
    id: number;
    from: string;
    description: string;
}

const acceptParams = (props: DialogProps): void => {
    dialogVisible.value = true;
    snapID.value = props.id;
    snapFrom.value = props.from;
    snapDescription.value = props.description;
    onWatch();
    nextTick(() => {
        loadCurrentStatus();
    });
};
const emit = defineEmits(['search']);

const loadCurrentStatus = async () => {
    loading.value = true;
    await loadSnapStatus(snapID.value)
        .then((res) => {
            loading.value = false;
            status.panel = res.data.panel;
            status.panelInfo = res.data.panelInfo;
            status.daemonJson = res.data.daemonJson;
            status.appData = res.data.appData;
            status.panelData = res.data.panelData;
            status.backupData = res.data.backupData;

            status.compress = res.data.compress;
            status.size = res.data.size;
            status.upload = res.data.upload;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onClose = async () => {
    emit('search');
    dialogVisible.value = false;
};

const onRetry = async () => {
    loading.value = true;
    await snapshotCreate({ id: snapID.value, from: snapFrom.value, description: snapDescription.value })
        .then(() => {
            loading.value = false;
            loadCurrentStatus();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onWatch = () => {
    timer = setInterval(async () => {
        if (keepLoadStatus()) {
            const res = await loadSnapStatus(snapID.value);
            status.panel = res.data.panel;
            status.panelInfo = res.data.panelInfo;
            status.daemonJson = res.data.daemonJson;
            status.appData = res.data.appData;
            status.panelData = res.data.panelData;
            status.backupData = res.data.backupData;

            status.compress = res.data.compress;
            status.size = res.data.size;
            status.upload = res.data.upload;
        }
    }, 1000 * 3);
};

const keepLoadStatus = () => {
    if (status.panel === 'Running') {
        return true;
    }
    if (status.panelInfo === 'Running') {
        return true;
    }
    if (status.daemonJson === 'Running') {
        return true;
    }
    if (status.appData === 'Running') {
        return true;
    }
    if (status.panelData === 'Running') {
        return true;
    }
    if (status.backupData === 'Running') {
        return true;
    }
    if (status.compress === 'Running') {
        return true;
    }
    if (status.upload === 'Uploading') {
        return true;
    }
    return false;
};

const showErrorMsg = (status: string) => {
    return status !== 'Running' && status !== 'Done' && status !== 'Uploading' && status !== 'Waiting';
};

const showRetry = () => {
    if (keepLoadStatus()) {
        return false;
    }
    if (status.panel !== 'Running' && status.panel !== 'Done') {
        return true;
    }
    if (status.panelInfo !== 'Running' && status.panelInfo !== 'Done') {
        return true;
    }
    if (status.daemonJson !== 'Running' && status.daemonJson !== 'Done') {
        return true;
    }
    if (status.appData !== 'Running' && status.appData !== 'Done') {
        return true;
    }
    if (status.panelData !== 'Running' && status.panelData !== 'Done') {
        return true;
    }
    if (status.backupData !== 'Running' && status.backupData !== 'Done') {
        return true;
    }
    if (status.compress !== 'Running' && status.compress !== 'Done' && status.compress !== 'Waiting') {
        return true;
    }
    if (status.upload !== 'Uploading' && status.upload !== 'Done' && status.upload !== 'Waiting') {
        return true;
    }
    return false;
};

const loadStatus = (status: string) => {
    switch (status) {
        case 'Running':
        case 'Waiting':
        case 'Uploading':
            return 'info';
        case 'Done':
            return 'success';
        default:
            return 'error';
    }
};

const loadIcon = (status: string) => {
    switch (status) {
        case 'Running':
        case 'Waiting':
        case 'Uploading':
            return 'Loading';
        case 'Done':
            return 'Check';
        default:
            return 'Close';
    }
};

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});
defineExpose({
    acceptParams,
});
</script>
<style scoped lang="scss">
.el-alert {
    margin: 10px 0 0;
}
.el-alert:first-child {
    margin: 0;
}
.top-margin {
    margin-top: 10px;
}
.err-message {
    margin-left: 23px;
    line-height: 20px;
    word-break: break-all;
    word-wrap: break-word;
}
</style>
