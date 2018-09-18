using System.Collections;
using System.Collections.Generic;
using UnityEngine;

using UnityEngine.UI;

public class ButtonScript : MonoBehaviour {
    public Button button1;
    public MexGrpcSample mexGrpcSample;

    void Start () {
        Button btn = button1.GetComponent<Button>();
        btn.onClick.AddListener(TaskOnClick);
    }

    void TaskOnClick(){
        Debug.Log ("You have clicked the button!");
        mexGrpcSample.RunSampleFlow();
    }

}
