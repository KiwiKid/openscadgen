include <BOSL2/std.scad>;

$fn = 100;
playerName = "Greg";
globalTextGap = 14;
/*
text3d();
    back(globalTextGap)
    text3d(str(""));
    right(globalTextGap)
    text3d();*/

    textLeft = 42;
    textDown = 26;
    
    mode = "text";
        globalGoalScale = [0.5,0.5,0.5];

    font_size = 7*globalGoalScale[0];
line_spacing = 18*globalGoalScale[0]; 
extrusion_height = 1;
    
    playerNameLength = 30;
   
    
text_lines =  ["Participant 2026",playerName,"Bailey Boys"];
text_lines_length = [80,playerNameLength, 60];
    total_height = (len(text_lines) - 1) * line_spacing;
    
   backPlateDown = 1;
      backPlateBack = 3;

    module textPlates(plateHeight=1.3){
        for (i = [0 : len(text_lines) - 1]) {
            y_offset = total_height / 2 - i * line_spacing;
    rotate([0,180,90])
translate([-textLeft, y_offset-textDown-1, -0.6]){
      //  down(backPlateDown)
            back(backPlateBack)
            cuboid([text_lines_length[i]*globalGoalScale[0],14*globalGoalScale[0],plateHeight*globalGoalScale[0]], anchor=BOT, rounding=2*globalGoalScale[0], edges="Z");
            }
    
        }
    }
module textShape(includeBackPlate=false, textHeight=extrusion_height){


// Loop and center multiple text3d modules


for (i = [0 : len(text_lines) - 1]) {
    y_offset = total_height / 2 - i * line_spacing;
    
    rotate([0,180,90])
    // Move each line to its stacked Y position
    translate([-textLeft, y_offset-textDown+1, 1.1]){
        text3d(text_lines[i], h = extrusion_height, size = font_size, anchor=CENTER, font="Trajan Pro:style=Bold");
        
        if(includeBackPlate){
     //       down(backPlateDown)
            back(backPlateBack)
            cuboid([text_lines_length[i],14,1], anchor=BOT, rounding=5, edges="Z");
        }
        
        }
}
}

module football_goal_labelled(mode="all"){
if(mode == "all"){
echo("ALLLLLL");
difference(){
    textPlates();
    up(0.7)
        textShape(includeBackPlate=false);
        
        }
    difference(){
    
scale(globalGoalScale)
	import("./goal.stl");
  //  down(-1*globalGoalScale[0])
    textPlates(plateHeight=10);
    }
    } else if(mode == "text"){
    
  /*  left(100)
    #cuboid([10,10,10], anchor=BOT)*/
   //down(1)
   up(0.9)
    textShape(includeBackPlate=false, textHeight=0.3);
    }
    
}
football_goal_labelled(mode=mode);
