using UnityEngine;

using UnityEngine.UI;

public class ButtonScript : MonoBehaviour {
    public Button startGprcButton;
    public MexGrpcSample mexGrpcSample;

    void Start () {
        Button btn = startGprcButton.GetComponent<Button>();
        btn.onClick.AddListener(TaskOnClick);
    }

    void TaskOnClick(){
        Debug.Log ("You have clicked the button!");
        mexGrpcSample.RunSampleFlow();
    }

}
