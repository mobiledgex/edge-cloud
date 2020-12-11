package edgeproto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testStatusMsgs(t *testing.T, infoStatus, diffStatus *StatusInfo, inMsgs, outMsgs []string, msgCnt int) {
	infoStatus.Msgs = inMsgs
	UpdateStatusDiff(infoStatus, diffStatus)
	require.Equal(t, len(diffStatus.Msgs), len(outMsgs))
	require.Equal(t, diffStatus.MsgCount, uint32(msgCnt))
	for ii := 0; ii < len(diffStatus.Msgs); ii++ {
		require.Equal(t, diffStatus.Msgs[ii], outMsgs[ii])
	}
}

func TestUpdateStatusDiff(t *testing.T) {
	diffStatus := StatusInfo{}
	infoStatus := StatusInfo{}
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1"}, []string{"1"}, 1)
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1", "2"}, []string{"2"}, 2)
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1", "2"}, []string{}, 2)
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1", "2", "3"}, []string{"3"}, 3)
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1", "2"}, []string{}, 3)
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1", "2", "3", "4", "5"}, []string{"4", "5"}, 5)
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1", "2", "3"}, []string{}, 5)
	testStatusMsgs(t, &infoStatus, &diffStatus, []string{"1", "2", "3", "4", "5"}, []string{}, 5)
}
