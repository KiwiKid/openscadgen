

	include <BOSL2/std.scad>;
include <BOSL2/hinges.scad>;
	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;
    
    renderType = "leavers";
        holderHeight = 28;
 minusOffset = -64; 
  plusOffset = minusOffset+22;
    holderLength = 180;
 holderWeight = 10;
 
 leaverHeight = 120;
 leaverRadius = 2;
 
 hingeRadius = 7;
 hingeHeight = 10;
 
 leaverBox = [30,30,20];
 hingeSnapdiam = 8;

 leverSink = 5;
 leaverMove = [-66.5,holderWeight,leverSink];
 leaverOnlyMove = [5+hingeSnapdiam,0,0];
 
 leaverHoldSize = [14,6,20];
 holeMoveAdjust = [14,-9.3,0];
 
 crossBarMove = [-13,0,8];
crossBarSize =  [25,3,5];
 
 buttonPresserHeight = 20;
  leaverPoleMove = [0,0,0];
snapLockHeight  = 10;
hingeSocketThick = 4;
 
 module hinge(){
 
        rotate([90,0,0])
        #cyl(r=hingeRadius, h=hingeHeight, chamfer=hingeHeight/3);
 
 }
 
 module volume_lever_holder(){
 rotate([0,0,90])
        snap_socket(thick=hingeSocketThick, foldangle=60,snapdiam=hingeSnapdiam, $slop=0.2);
    //    cuboid(leaverBox);
 }
 
 module volume_lever(){
    //    hinge();
    
     right(13.8){

        rotate([0,0,90])
        snap_lock(thick=hingeSocketThick, foldangle=60, snapdiam=hingeSnapdiam);
        
              
     /*   up(snapLockHeight+1)
        rotate([0,180,90])
        snap_lock(thick=snapLockHeight, foldangle=60, snapdiam=hingeSnapdiam);*/
        
                    
                move(crossBarMove)
                 cuboid(crossBarSize);
        }
     up(leaverHeight/4)
     cyl(r=leaverRadius, h=leaverHeight/3);
     
     
     
     
 }
 
module button_holder(diameter, holderWeight, holderHeight = 25){

up(holderHeight/2)
down(holderWeight/2)
union(){
    cyl(d=diameter, h=holderHeight, center=true, chamfer1=-4, chamfer2=2, rounding=2);
  
    }
}

	module echo_show_15_volume_buttons(){
    
 
  
    back(holderHeight/3)
		difference(){
			cuboid([holderLength,holderHeight,holderWeight], rounding=2,  edges=[TOP]);
            down(4)
			cuboid([holderLength+2,holderHeight*0.8,holderWeight]);
         //   down(7)
          // cuboid([holderLength+2,holderHeight*1.5,holderWeight]);
            // swing_text_up("ALEXA - +");
            
         up(holderWeight/2*1.2-0.01)
            right(minusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=30);
            
          up(holderWeight/2*1.2-0.01)
            right(plusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=30);
            
            // leaver swing gap
        move(leaverMove){
        move(holeMoveAdjust)
            cuboid(leaverHoldSize);
            }
		}
        
        
        
        move(leaverMove){
        volume_lever_holder();
        }
        
        move(leaverMove){
        
        move(leaverOnlyMove)
        volume_lever();
        }
       
            // Top button labels
          /*  textWidthOffset = -3;
            back(3)
               right(plusOffset)
               translate([0, textWidthOffset, holderHeight-8.5])  
               linear_extrude(height=3)
               text("↑", size=26, halign="center", valign="bottom");
               
               right(minusOffset)
               translate([0, 10+textWidthOffset, holderHeight-8.5])
               linear_extrude(height=3)
               text("↓", size=26, halign="center", valign="bottom");
               */
           
           
           
           // Button pressing drivers
           union(){
            back(holderHeight/3)
           // back(holderWeight/2)
            right(minusOffset)
            button_holder(diameter=10.5, holderWeight=holderWeight, holderHeight=buttonPresserHeight);
            
            back(holderHeight/3)
          //  back(holderWeight/2)
            right(plusOffset)
            button_holder(diameter=10.5, holderWeight=holderWeight, holderHeight=buttonPresserHeight);
        
        }
        
           difference(){
        right(2.5)
            fwd(4)
                up(6)
                union(){
                right(30)
            swing_text_up("ALEXA", textSize=28);
            /*fwd(3)
            left(70.5)
             swing_text_up(" ▼", textSize=27);
             
             fwd(3)
                         left(50)
                          swing_text_up(" ▲", textSize=27);
*/
             
                                  
             
           
            }
        
    /*    
        #back(holderHeight/3)
            right(minusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=holderHeight);
            
            #back(holderHeight/3)
            right(plusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=holderHeight);
        }*/
        }
       }    
        
        module swing_text_up(text_str, textSize=20) {
    // Parameters
    text_height = 1;

            for (a = [-14 : 0.5 : 90]) {
                // Shift text so bottom edge is at origin (hinge line)
                translate([0, 0, 0])
                    rotate([a, 0, 0])
                        translate([0, 0, 0])  // pivot point at bottom edge
                            linear_extrude(height=text_height)
                                text(text_str, size=textSize, halign="center", valign="bottom");
            }
        }
        
        
        module text_board(text_str, textSize=20){
            linear_extrude(height=text_height)
                                    text(text_str, size=textSize, halign="center", valign="bottom");
         }
        
     
        
        


	


    sliced(renderType=renderType) {
        if(renderType == "leavers"){
        intersection(){
        
            echo_show_15_volume_buttons();
            
            left(50)
            cuboid([40,100,100]);
            }

        }else {
        echo_show_15_volume_buttons();
        }
    }
    
        
    bracketBaseMove = 15;
   bracketGap = 5;
        
    module volumeSymbol(){
    text_board(" ◄", textSize=27);
    
    down(3){
        
        right(bracketBaseMove){
        
                    right(bracketGap){
            text_board(")", textSize=20);
            }
            
            right(bracketGap*2){
            text_board(")", textSize=20);
            }
            
            right(bracketGap*3){
            text_board(")", textSize=20);
            }
            }
        }
    }

    left(200)
volumeSymbol();






	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.3,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 7],
    vertSlicePos = [-3, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
        rotate([0,90,0])
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                rotate([0,90,0])
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
        rotate([0,0,90])
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
         
            intersection() {
                children();
                  rotate([0,0,90])
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

