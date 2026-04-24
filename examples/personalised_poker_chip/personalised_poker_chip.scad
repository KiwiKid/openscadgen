

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
chipRadius = 19;
chipInnerRadius = 15;
chipInnerRadiusOuter = 9.5;
chipInnerRadiusInner = 8;
//chipHeight = .5;
//chipInnerHeight = 1.2;

chipHeight = 1.5;
chipInnerHeight = 0.9;
innerSlotUp = 0.8;



 
textUp = chipHeight+0.2;
textDown = 0.20;

centerText = "4";
centerTextSize = 14;
centerTextHeight = 0.8;


bottomText = "PEG";
bottomTextSize = 6;
 roundLetterSize  = 7.5;
 bottomTitleAngleStart = 180;
 textBottomRadius = 16.5;

topText = "BAILEY BOYS";
textRadius = 12.0;
topRoundLetter = 4.0;
topRoundLetterSize = 4;
 topTitleAngleStart = -200;
 
underSideText = "2026";//"Got your chip?";
underSideTextSize = 6;
underSideTextLettersize = 6;
textUndersideRadius = 10.2;

 outerToothOffset = 0.5;
 font = "Arial Rounded MT Bold:style=Regular";
 
 module pipe(){
difference(){
     cyl(r=chipInnerRadiusOuter, h=chipHeight, anchor=BOTTOM, chamfer1=0.3,  chamfer2=-0.3);
       cyl(r=chipInnerRadiusInner, h=chipInnerHeight+3, anchor=BOTTOM, chamfer1=-0.2);     }
 }
 
module tooth() {
    cuboid([0.4,2, 3], anchor=FRONT);
//    rect([0.4, 1.5], anchor=FRONT);
}

module rotatedElements() {
    r = chipRadius - 1;

    // optional outline
   // color("black")
      //  circle(r = r, $fn = 64);

    // copies around the chip
   // color("red")
    for (a = [0 : 360/100 : 360 - 360/100]) {
    rotate(a)
        translate([0, r+outerToothOffset, 1])
            tooth();
    }
}

module bottomDecoration(){
     for (a = [0  : 30 : 360]) {
    rotate([0, 0,a])
      // translate([0, a+outerToothOffset, 1])
      
     translate([4.5,0,0.0])
            cuboid([7,0.5,0.8], chamfer=0.4, edges="Y");
    }
}

	module personalised_poker_chip(){
    difference(){
		cyl(r=chipRadius, h=chipHeight, anchor=BOTTOM,chamfer=0.18);
        
        up(innerSlotUp)
            
                   pipe();
                   
                   
           up(chipHeight)
           text3d(centerText, font=font, center=true, size=centerTextSize, h=centerTextHeight);
           
          rotatedElements();
          
          bottomDecoration();
        //[220, -40]
             path = path3d(arc(80, r=textRadius, angle=[220, -40]));

                
        //     up(textUp)
      //  color("red")
       // stroke(path, width=.3);
        
        //rotate([180,0,0])
             up(textUp)
        path_text(path, topText, font=font, size=topRoundLetterSize,  lettersize = topRoundLetter, normal=UP, center=true);
                
               
      /*         #stroke(path,  width=.3);
        //rotate([180,0,0])
             up(textUp)
        path_text(path, underSideText, font=font, size=underSideTextSize,  lettersize = underSideTextLettersize, normal=UP, center=true);
        */
        //debug
      //  # stroke(path, width=0.5);
        
        
       path2 = path3d(arc(120,r=textBottomRadius, angle=[bottomTitleAngleStart, 360]), );
    //   up(10)
       
      //   color("red")
       //  #stroke(path2, width=0.5);
         
        up(textUp)
       path_text(path2, bottomText, size=bottomTextSize, font=font, lettersize = roundLetterSize, normal=UP, center=true);
        
        //debug
       // stroke(path2, width=0.5);
        
        if(underSideText != ""){
        path3 = path3d(arc(200,r=textUndersideRadius, angle=[0,180]), );
       // color("red")
       //  stroke(path, width=1);
        //up(textUp)
        down(textDown)
        path_text(path3, underSideText, size=underSideTextSize, font=font, lettersize = underSideTextLettersize, normal=DOWN, center=true);
        }
}
	}
    
    alignBlockSize = [1,10,1];
    alignOffset = 10;
    module alignBlock(){
        left(alignOffset)
        cuboid(alignBlockSize, anchor=CENTER);
        
        right(alignOffset)
        cuboid(alignBlockSize, anchor=CENTER);
        }
    
    
    if(renderType == "obj"){
        personalised_poker_chip();
    }
    if(renderType == "backHalf"){
        difference(){
        
        bottom_half()
        down(chipHeight/2)
            personalised_poker_chip();
            alignBlock();
            }
            
            

    }
    if(renderType == "frontHalf"){
    
            difference(){
        
        top_half()
        down(chipHeight/2)
            personalised_poker_chip();
            alignBlock();
            }

    }
    
    if(renderType == "alignBlock"){
        alignBlock();
        }
    
    
       









